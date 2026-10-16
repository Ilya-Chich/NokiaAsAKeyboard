package e70kbd;

import javax.microedition.rms.RecordStore;
import javax.microedition.rms.RecordStoreException;

/**
 * One RMS record: version, 6-byte Bluetooth address, big-endian channel.
 * No socket is opened here. Channel stays at the stored integer until a
 * later step writes the last working RFCOMM channel.
 */
final class Store {

    static final String STORE_NAME = "e70kbd";

    private static final int VERSION = 1;
    private static final int RECORD_LENGTH = 11;

    /** 12 uppercase hex digits, or null when unset. */
    String address;

    /** RFCOMM channel. 1 until a later step learns a working channel. */
    int channel = 1;

    void load() throws RecordStoreException {
        RecordStore rs = RecordStore.openRecordStore(STORE_NAME, true);
        try {
            if (rs.getNumRecords() < 1) {
                return;
            }
            byte[] rec = rs.getRecord(1);
            if (rec == null || rec.length < RECORD_LENGTH || rec[0] != VERSION) {
                return;
            }
            StringBuffer hex = new StringBuffer(12);
            for (int i = 0; i < 6; i++) {
                int b = rec[1 + i] & 0xff;
                hex.append(hexDigit(b >>> 4));
                hex.append(hexDigit(b & 0x0f));
            }
            address = hex.toString();
            channel = ((rec[7] & 0xff) << 24)
                    | ((rec[8] & 0xff) << 16)
                    | ((rec[9] & 0xff) << 8)
                    | (rec[10] & 0xff);
        } finally {
            rs.closeRecordStore();
        }
    }

    void save() throws RecordStoreException {
        byte[] rec = new byte[RECORD_LENGTH];
        rec[0] = (byte) VERSION;
        for (int i = 0; i < 6; i++) {
            int hi = Character.digit(address.charAt(i * 2), 16);
            int lo = Character.digit(address.charAt(i * 2 + 1), 16);
            rec[1 + i] = (byte) ((hi << 4) | lo);
        }
        rec[7] = (byte) (channel >>> 24);
        rec[8] = (byte) (channel >>> 16);
        rec[9] = (byte) (channel >>> 8);
        rec[10] = (byte) channel;

        RecordStore rs = RecordStore.openRecordStore(STORE_NAME, true);
        try {
            if (rs.getNumRecords() < 1) {
                rs.addRecord(rec, 0, rec.length);
            } else {
                rs.setRecord(1, rec, 0, rec.length);
            }
        } finally {
            rs.closeRecordStore();
        }
    }

    /** 12 hex digits, colons optional. Anything else is rejected. */
    static String normalize(String raw) {
        if (raw == null) {
            return null;
        }
        StringBuffer hex = new StringBuffer(12);
        for (int i = 0; i < raw.length(); i++) {
            char c = raw.charAt(i);
            if (c == ':') {
                continue;
            }
            int d = Character.digit(c, 16);
            if (d < 0 || hex.length() == 12) {
                return null;
            }
            hex.append(hexDigit(d));
        }
        if (hex.length() != 12) {
            return null;
        }
        return hex.toString();
    }

    static String format(String hex12) {
        StringBuffer out = new StringBuffer(17);
        for (int i = 0; i < 12; i++) {
            if (i > 0 && (i & 1) == 0) {
                out.append(':');
            }
            out.append(hex12.charAt(i));
        }
        return out.toString();
    }

    private static char hexDigit(int n) {
        if (n < 10) {
            return (char) ('0' + n);
        }
        return (char) ('A' + (n - 10));
    }
}
