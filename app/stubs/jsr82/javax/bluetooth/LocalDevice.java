/*
 * Compile-only JSR-82 stub. The phone has the real class. Never packaged.
 */
package javax.bluetooth;

public class LocalDevice {
    public static final int GIAC = 0x9E8B33;
    public static final int LIAC = 0x9E8B00;
    public static final int NOT_DISCOVERABLE = 0x00;

    public static LocalDevice getLocalDevice() throws BluetoothStateException {
        return null;
    }

    public DiscoveryAgent getDiscoveryAgent() {
        return null;
    }

    public String getBluetoothAddress() {
        return null;
    }

    public String getFriendlyName() {
        return null;
    }

    public DeviceClass getDeviceClass() {
        return null;
    }

    public boolean setDiscoverable(int mode) throws BluetoothStateException {
        return false;
    }

    public int getDiscoverable() {
        return 0;
    }

    public static String getProperty(String property) {
        return null;
    }
}
