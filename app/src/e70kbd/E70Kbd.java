package e70kbd;

import javax.microedition.lcdui.Canvas;
import javax.microedition.lcdui.Command;
import javax.microedition.lcdui.CommandListener;
import javax.microedition.lcdui.Display;
import javax.microedition.lcdui.Displayable;
import javax.microedition.lcdui.Font;
import javax.microedition.lcdui.Form;
import javax.microedition.lcdui.Graphics;
import javax.microedition.midlet.MIDlet;

public class E70Kbd extends MIDlet {

    private Dump dump;
    private Store store;
    private AddrForm form;
    private Link link;

    public void startApp() {
        Display display = Display.getDisplay(this);
        if (store == null) {
            store = new Store();
            try {
                store.load();
            } catch (Throwable t) {
                showError(display, t);
                return;
            }
        }
        if (dump == null) {
            dump = new Dump(this, store);
        }
        if (link == null) {
            link = new Link(dump, store);
            dump.setLink(link);
            link.start();
        }
        try {
            display.setCurrent(dump);
            if (store.address == null) {
                display.callSerially(new Runnable() {
                    public void run() {
                        showAddress();
                    }
                });
            }
        } catch (Throwable t) {
            showError(display, t);
        }
    }

    public void pauseApp() {
    }

    public void destroyApp(boolean unconditional) {
        if (link != null) {
            link.stop();
        }
    }

    void showAddress() {
        Display display = Display.getDisplay(this);
        if (form == null) {
            form = new AddrForm(this, store);
        }
        form.prepare();
        display.setCurrent(form);
        display.callSerially(form);
    }

    void showDump() {
        Display.getDisplay(this).setCurrent(dump);
    }

    private static void showError(Display display, Throwable t) {
        Form error = new Form("E70Kbd");
        error.append(t.toString());
        display.setCurrent(error);
    }
}

/**
 * Shows keyPressed, keyReleased, and keyRepeated.
 * Commands are added in showNotify. The E70 aborts launch if they are
 * added in the constructor, before the canvas is current.
 */
final class Dump extends Canvas implements CommandListener {

    private static final int LINES = 16;

    private final E70Kbd midlet;
    private final Store store;
    private final String[] lines = new String[LINES];
    private int count;
    private boolean commandsAdded;
    private Command addressCommand;
    private Command exitCommand;
    private String status = "idle";
    private Link link;

    Dump(E70Kbd midlet, Store store) {
        this.midlet = midlet;
        this.store = store;
    }

    void setLink(Link link) {
        this.link = link;
    }

    void setStatus(String text) {
        status = text;
        repaint();
    }

    protected void showNotify() {
        if (!commandsAdded) {
            addressCommand = new Command("Address", Command.OK, 1);
            exitCommand = new Command("Exit", Command.EXIT, 2);
            addCommand(addressCommand);
            addCommand(exitCommand);
            setCommandListener(this);
            commandsAdded = true;
        }
        if (link != null) {
            link.canvasShown();
        }
    }

    protected void hideNotify() {
        if (link != null) {
            link.canvasHidden();
        }
    }

    protected void keyPressed(int key) {
        add("pressed", key);
        if (link != null) {
            link.tap(key);
        }
    }

    protected void keyReleased(int key) {
        add("released", key);
    }

    protected void keyRepeated(int key) {
        add("repeated", key);
    }

    private void add(String event, int key) {
        String name = "";
        try {
            name = getKeyName(key);
        } catch (IllegalArgumentException e) {
            name = "";
        }
        String line = event + " " + key;
        if (name != null && name.length() > 0) {
            line = line + " " + name;
        }
        if (count < LINES) {
            lines[count] = line;
            count++;
        } else {
            System.arraycopy(lines, 1, lines, 0, LINES - 1);
            lines[LINES - 1] = line;
        }
        repaint();
    }

    protected void paint(Graphics g) {
        int w = getWidth();
        int h = getHeight();
        g.setColor(0xFFFFFF);
        g.fillRect(0, 0, w, h);
        g.setColor(0);
        g.setFont(Font.getFont(Font.FACE_SYSTEM, Font.STYLE_PLAIN, Font.SIZE_MEDIUM));
        int lh = g.getFont().getHeight();
        if (lh < 1) {
            lh = 1;
        }
        int y = 0;
        String shown = "no address";
        if (store.address != null) {
            shown = Store.format(store.address);
        }
        g.drawString(shown, 1, y, Graphics.TOP | Graphics.LEFT);
        y += lh;
        y = drawWrapped(g, status, 1, y, w - 2, lh, h);
        int n = count;
        int visible = (h - y) / lh;
        if (visible < 1) {
            visible = 1;
        }
        int start = 0;
        if (n > visible) {
            start = n - visible;
        }
        for (int i = start; i < n; i++) {
            g.drawString(lines[i], 1, y, Graphics.TOP | Graphics.LEFT);
            y += lh;
        }
    }

    private static int drawWrapped(Graphics g, String text, int x, int y, int width, int lh, int limit) {
        if (text == null || text.length() == 0) {
            return y + lh;
        }
        Font font = g.getFont();
        int start = 0;
        while (start < text.length() && y + lh <= limit) {
            int end = start + 1;
            while (end < text.length() && font.stringWidth(text.substring(start, end + 1)) <= width) {
                end++;
            }
            g.drawString(text.substring(start, end), x, y, Graphics.TOP | Graphics.LEFT);
            y += lh;
            start = end;
        }
        return y;
    }

    public void commandAction(Command command, Displayable displayable) {
        if (command == addressCommand) {
            midlet.showAddress();
            return;
        }
        if (command == exitCommand) {
            midlet.notifyDestroyed();
        }
    }
}
