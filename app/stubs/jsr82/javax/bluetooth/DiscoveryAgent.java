/*
 * Compile-only JSR-82 stub. The phone has the real class. Never packaged.
 */
package javax.bluetooth;

public class DiscoveryAgent {
    public static final int NOT_DISCOVERABLE = 0x00;
    public static final int GIAC = 0x9E8B33;
    public static final int LIAC = 0x9E8B00;
    public static final int CACHED = 0x00;
    public static final int PREKNOWN = 0x01;

    public RemoteDevice[] retrieveDevices(int option) {
        return null;
    }

    public boolean startInquiry(int accessCode, DiscoveryListener listener)
            throws BluetoothStateException {
        return false;
    }

    public boolean cancelInquiry(DiscoveryListener listener) {
        return false;
    }

    public int searchServices(int[] attrSet, UUID[] uuidSet, RemoteDevice btDev,
            DiscoveryListener discListener) throws BluetoothStateException {
        return 0;
    }

    public boolean cancelServiceSearch(int transID) {
        return false;
    }

    public String selectService(UUID uuid, int security, boolean master)
            throws BluetoothStateException {
        return null;
    }
}
