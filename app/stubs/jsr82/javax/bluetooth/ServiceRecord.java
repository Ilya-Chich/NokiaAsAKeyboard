/*
 * Compile-only JSR-82 stub. The phone has the real class. Never packaged.
 */
package javax.bluetooth;

import java.io.IOException;

public interface ServiceRecord {
    int NOAUTHENTICATE_NOENCRYPT = 0x00;
    int AUTHENTICATE_NOENCRYPT = 0x01;
    int AUTHENTICATE_ENCRYPT = 0x02;

    int[] getAttributeIDs();

    DataElement getAttributeValue(int attrID);

    RemoteDevice getHostDevice();

    boolean populateRecord(int[] attrIDs) throws IOException;

    String getConnectionURL(int requiredSecurity, boolean mustBeMaster);

    void setAttributeValue(int attrID, DataElement attrValue);

    void setDeviceServiceClasses(int classes);
}
