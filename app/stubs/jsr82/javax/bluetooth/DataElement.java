/*
 * Compile-only JSR-82 stub. The phone has the real class. Never packaged.
 */
package javax.bluetooth;

import java.util.Enumeration;

public class DataElement {
    public static final int NULL = 0;
    public static final int U_INT_1 = 0x08;
    public static final int U_INT_2 = 0x09;
    public static final int U_INT_4 = 0x0A;
    public static final int U_INT_8 = 0x0B;
    public static final int U_INT_16 = 0x0C;
    public static final int INT_1 = 0x10;
    public static final int INT_2 = 0x11;
    public static final int INT_4 = 0x12;
    public static final int INT_8 = 0x13;
    public static final int INT_16 = 0x14;
    public static final int URL = 0x40;
    public static final int UUID = 0x18;
    public static final int BOOL = 0x28;
    public static final int STRING = 0x20;
    public static final int DATSEQ = 0x30;
    public static final int DATALT = 0x38;

    public DataElement(int valueType) {
    }

    public DataElement(boolean bool) {
    }

    public DataElement(int valueType, long value) {
    }

    public DataElement(int valueType, Object value) {
    }

    public void addElement(DataElement elem) {
    }

    public void insertElementAt(DataElement elem, int index) {
    }

    public boolean removeElement(DataElement elem) {
        return false;
    }

    public int getDataType() {
        return 0;
    }

    public long getLong() {
        return 0;
    }

    public boolean getBoolean() {
        return false;
    }

    public int getSize() {
        return 0;
    }

    public Object getValue() {
        return null;
    }

    public Enumeration getElements() {
        return null;
    }
}
