/*
 * Compile-only JSR-82 stub. The phone has the real class. Never packaged.
 */
package javax.bluetooth;

import java.io.IOException;

import javax.microedition.io.Connection;

public class RemoteDevice {
    protected RemoteDevice(String address) {
    }

    public static RemoteDevice getRemoteDevice(Connection conn) throws IOException {
        return null;
    }

    public String getBluetoothAddress() {
        return null;
    }

    public String getFriendlyName(boolean alwaysAsk) throws IOException {
        return null;
    }

    public boolean authenticate() throws IOException {
        return false;
    }

    public boolean authorize(Connection conn) throws IOException {
        return false;
    }

    public boolean encrypt(Connection conn, boolean on) throws IOException {
        return false;
    }

    public boolean isAuthenticated() {
        return false;
    }

    public boolean isAuthorized(Connection conn) throws IOException {
        return false;
    }

    public boolean isEncrypted() {
        return false;
    }

    public boolean isTrustedDevice() {
        return false;
    }
}
