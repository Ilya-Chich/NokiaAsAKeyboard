package e70kbd;

import javax.microedition.lcdui.Command;
import javax.microedition.lcdui.CommandListener;
import javax.microedition.lcdui.Displayable;
import javax.microedition.lcdui.Form;
import javax.microedition.lcdui.StringItem;
import javax.microedition.lcdui.TextField;

/**
 * Address editor. Commands are added from run(), after the form is current.
 * The E70 aborts launch if commands are added before the first screen is shown.
 */
final class AddrForm extends Form implements CommandListener, Runnable {

    private final E70Kbd midlet;
    private final Store store;
    private final TextField field;
    private final StringItem message;
    private Command saveCommand;
    private Command backCommand;
    private boolean commandsAdded;

    AddrForm(E70Kbd midlet, Store store) {
        super("PC address");
        this.midlet = midlet;
        this.store = store;
        field = new TextField("Address", "", 17, TextField.ANY);
        message = new StringItem(null, "12 hex digits, colons optional");
        append(field);
        append(message);
    }

    void prepare() {
        if (store.address == null) {
            field.setString("");
        } else {
            field.setString(Store.format(store.address));
        }
        message.setText("12 hex digits, colons optional");
    }

    public void run() {
        if (commandsAdded) {
            return;
        }
        saveCommand = new Command("Save", Command.OK, 1);
        backCommand = new Command("Back", Command.BACK, 2);
        addCommand(saveCommand);
        addCommand(backCommand);
        setCommandListener(this);
        commandsAdded = true;
    }

    public void commandAction(Command command, Displayable displayable) {
        if (command == backCommand) {
            midlet.showDump();
            return;
        }
        if (command != saveCommand) {
            return;
        }
        String hex = Store.normalize(field.getString());
        if (hex == null) {
            message.setText("need 12 hex digits");
            return;
        }
        String previous = store.address;
        int previousChannel = store.channel;
        store.address = hex;
        try {
            store.save();
        } catch (Throwable t) {
            store.address = previous;
            store.channel = previousChannel;
            message.setText(t.toString());
            return;
        }
        midlet.showDump();
    }
}
