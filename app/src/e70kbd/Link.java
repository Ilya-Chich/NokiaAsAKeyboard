package e70kbd;

import java.io.DataInputStream;
import java.io.DataOutputStream;
import java.io.IOException;
import java.util.Vector;

import javax.bluetooth.BluetoothStateException;
import javax.bluetooth.DeviceClass;
import javax.bluetooth.DiscoveryAgent;
import javax.bluetooth.DiscoveryListener;
import javax.bluetooth.LocalDevice;
import javax.bluetooth.RemoteDevice;
import javax.bluetooth.ServiceRecord;
import javax.bluetooth.UUID;
import javax.microedition.io.Connector;
import javax.microedition.io.StreamConnection;

/**
 * Dials the saved PC while the canvas is showing. HELLO, then PING.
 * No key is sent.
 */
final class Link implements Runnable {

    private static final String SERVICE = "8b0d1a3e70e74c1a9b5e6f0a0e70c0de";
    private static final byte[] HELLO = new byte[] {0x4B, 0x42, 1, 1, 1, 1};
    private static final byte[] PING = new byte[] {0x4B, 0x42, 1, 3, 0};
    private static final int SEARCH_MS = 8000;
    private static final int RETRY_MS = 3000;
    private static final int PING_MS = 5000;
    /** Same fixed channel the PC daemon requests from BlueZ. */
    private static final int DAEMON_CHANNEL = 22;

    private final Dump dump;
    private final Store store;
    private final Object connLock = new Object();
    private final Vector pending = new Vector();
    private volatile boolean alive = true;
    private volatile boolean canvasUp;
    private volatile boolean broken;
    private StreamConnection conn;

    Link(Dump dump, Store store) {
        this.dump = dump;
        this.store = store;
    }

    void start() {
        Thread thread = new Thread(this);
        thread.start();
    }

    void canvasShown() {
        canvasUp = true;
    }

    void canvasHidden() {
        canvasUp = false;
        drop();
    }

    void stop() {
        alive = false;
        canvasUp = false;
        drop();
    }

    void tap(int key) {
        byte[] frame = frameFor(key);
        if (frame == null) {
            return;
        }
        synchronized (pending) {
            if (pending.size() >= 32) {
                return;
            }
            pending.addElement(frame);
        }
    }

    public void run() {
        while (alive) {
            if (!canvasUp || store.address == null) {
                if (store.address == null) {
                    dump.setStatus("no address");
                } else {
                    dump.setStatus("paused");
                }
                pause(300);
                continue;
            }
            try {
                session();
            } catch (Throwable t) {
                dump.setStatus(BtUtil.text(t));
            } finally {
                drop();
            }
            if (alive && canvasUp) {
                pause(RETRY_MS);
            }
        }
    }

    private void session() throws IOException {
        broken = false;
        String url = findUrl();
        dump.setStatus("connecting " + BtUtil.channelOf(url));
        StreamConnection opened = (StreamConnection) Connector.open(url);
        synchronized (connLock) {
            if (!alive || !canvasUp) {
                opened.close();
                return;
            }
            conn = opened;
        }
        DataOutputStream out = opened.openDataOutputStream();
        DataInputStream in = opened.openDataInputStream();
        Thread reader = new Thread(new Reader(in));
        reader.start();
        out.write(HELLO);
        out.flush();
        dump.setStatus("up");
        long next = System.currentTimeMillis() + PING_MS;
        while (alive && canvasUp && !broken) {
            drain(out);
            long now = System.currentTimeMillis();
            if (now >= next) {
                out.write(PING);
                out.flush();
                next = now + PING_MS;
            }
            pause(50);
        }
    }

    private void drain(DataOutputStream out) throws IOException {
        for (;;) {
            byte[] frame;
            synchronized (pending) {
                if (pending.isEmpty()) {
                    return;
                }
                frame = (byte[]) pending.elementAt(0);
                pending.removeElementAt(0);
            }
            out.write(frame);
            out.flush();
        }
    }

    private static byte[] frameFor(int key) {
        if (key == 10) {
            return special((byte) 1);
        }
        if (key == 8) {
            return special((byte) 2);
        }
        if (key < 32 || key > 126) {
            return null;
        }
        if ("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 !\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~".indexOf((char) key) < 0) {
            return null;
        }
        return new byte[] {0x4B, 0x42, 1, 2, 2, 0, (byte) key};
    }

    private static byte[] special(byte code) {
        return new byte[] {0x4B, 0x42, 1, 2, 2, 1, code};
    }

    private String findUrl() {
        String found = null;
        try {
            found = search();
        } catch (Throwable t) {
            dump.setStatus(BtUtil.text(t));
        }
        if (found != null) {
            remember(found);
            return found;
        }
        int channel = store.channel;
        if (channel <= 1) {
            channel = DAEMON_CHANNEL;
        }
        return "btspp://" + store.address + ":" + channel
                + ";authenticate=true;encrypt=true";
    }

    private String search() throws BluetoothStateException, IOException {
        LocalDevice local = LocalDevice.getLocalDevice();
        DiscoveryAgent agent = local.getDiscoveryAgent();
        RemoteDevice device = BtUtil.findDevice(agent, store.address);
        if (device == null) {
            return null;
        }
        Search search = new Search();
        UUID[] uuids = new UUID[] {new UUID(SERVICE, false)};
        agent.searchServices(null, uuids, device, search);
        synchronized (search) {
            if (!search.done) {
                try {
                    search.wait(SEARCH_MS);
                } catch (InterruptedException e) {
                    return null;
                }
            }
        }
        if (search.url == null) {
            return null;
        }
        return BtUtil.secure(search.url);
    }

    private void remember(String url) {
        int channel = BtUtil.channelOf(url);
        if (channel < 1 || channel == store.channel) {
            return;
        }
        store.channel = channel;
        try {
            store.save();
        } catch (Throwable t) {
            dump.setStatus(BtUtil.text(t));
        }
    }

    private void drop() {
        synchronized (connLock) {
            if (conn != null) {
                try {
                    conn.close();
                } catch (IOException e) {
                    // The reader and writer both observe the closed socket.
                }
                conn = null;
            }
        }
    }

    private void pause(int ms) {
        int left = ms;
        while (left > 0 && alive) {
            int step = 200;
            if (step > left) {
                step = left;
            }
            try {
                Thread.sleep(step);
            } catch (InterruptedException e) {
                return;
            }
            left -= step;
        }
    }

    private final class Reader implements Runnable {
        private final DataInputStream in;

        Reader(DataInputStream in) {
            this.in = in;
        }

        public void run() {
            byte[] buf = new byte[32];
            try {
                while (in.read(buf) >= 0) {
                    // PING replies must be consumed. They are not shown.
                }
            } catch (IOException e) {
                broken = true;
            }
        }
    }
}

final class Search implements DiscoveryListener {
    String url;
    boolean done;

    public void deviceDiscovered(RemoteDevice btDevice, DeviceClass cod) {
    }

    public void inquiryCompleted(int discType) {
    }

    public void servicesDiscovered(int transID, ServiceRecord[] recs) {
        if (recs == null) {
            return;
        }
        for (int i = 0; i < recs.length; i++) {
            String found = recs[i].getConnectionURL(ServiceRecord.AUTHENTICATE_ENCRYPT, false);
            if (found != null) {
                url = found;
                return;
            }
        }
    }

    public void serviceSearchCompleted(int transID, int respCode) {
        synchronized (this) {
            done = true;
            notifyAll();
        }
    }
}

final class BtUtil {
    private BtUtil() {
    }

    static RemoteDevice findDevice(DiscoveryAgent agent, String address) {
        RemoteDevice found = match(agent.retrieveDevices(DiscoveryAgent.PREKNOWN), address);
        if (found != null) {
            return found;
        }
        return match(agent.retrieveDevices(DiscoveryAgent.CACHED), address);
    }

    private static RemoteDevice match(RemoteDevice[] devices, String address) {
        if (devices == null) {
            return null;
        }
        for (int i = 0; i < devices.length; i++) {
            String got = devices[i].getBluetoothAddress();
            if (got != null && got.equalsIgnoreCase(address)) {
                return devices[i];
            }
        }
        return null;
    }

    static String secure(String url) {
        String out = url;
        if (out.indexOf("authenticate=") < 0) {
            out = out + ";authenticate=true";
        }
        if (out.indexOf("encrypt=") < 0) {
            out = out + ";encrypt=true";
        }
        return out;
    }

    static int channelOf(String url) {
        int scheme = url.indexOf("://");
        if (scheme < 0) {
            return -1;
        }
        int colon = url.indexOf(':', scheme + 3);
        if (colon < 0 || colon + 1 >= url.length()) {
            return -1;
        }
        int value = 0;
        int i = colon + 1;
        if (url.charAt(i) < '0' || url.charAt(i) > '9') {
            return -1;
        }
        while (i < url.length()) {
            char c = url.charAt(i);
            if (c < '0' || c > '9') {
                break;
            }
            value = value * 10 + (c - '0');
            i++;
        }
        return value;
    }

    static String text(Throwable t) {
        String s = t.toString();
        if (s.startsWith("java.io.")) {
            s = s.substring(8);
        }
        return s;
    }
}
