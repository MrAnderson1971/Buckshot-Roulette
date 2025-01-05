package src;

import java.io.PrintWriter;
import java.util.Random;

import static src.BuckshotRoulette.*;

public abstract class Item {
    public abstract String text();
    public abstract void use(PrintWriter out, String player);

    public static class MagnifyingGlass extends Item {
        public String text() {
            return "🔍 Reveals the next shell.";
        }
        public void use(PrintWriter out, String player) {
            System.out.println("The next item is a " + shells.get(0).toString() + " shell.\n");
            sendMessage(out, "summary:", "Opponent used magnifying glass (very interesting)...\n");
        }
    }

    public static class Cigarette extends Item {
        public String text() {
            return "🚬 Restore one HP.";
        }
        public void use(PrintWriter out, String player) {
            hp.put(player, hp.get(player) + 1);
            System.out.println("Smoked one HP back.\n");
            sendMessage(out, "heal:", player + ",1,Opponent smoked 1 HP.\n");
        }
    }

    public static class Handsaw extends Item {
        public String text() {
            return "🪚 Next shot does double damage.";
        }
        public void use(PrintWriter out, String player) {
            damage = 2;
            System.out.println("Sawed off shotgun...\n");
            sendMessage(out, "summary:", "Opponent used handsaw...\n");
        }
    }

    public static class Beer extends Item {
        public String text() {
            return "🍺 Ejects the current shell.";
        }
        public void use(PrintWriter out, String player) {
            Shell first = shells.get(0);
            shells.remove(0);
            System.out.println("Ejected a " + first.toString() + " shell.\n");
            sendMessage(out, "eject:", "Opponent ejected a " + first + " shell.\n");
        }
    }

    public static class Handcuffs extends Item {
        public String text() {
            return "🔗 Skips your opponent's turn.";
        }
        public void use(PrintWriter out, String player) {
            cuffedOpponent = true;
            System.out.println("Cuffed your opponent.");
            sendMessage(out, "summary:", "Opponent cuffed you!\n");
        }
    }

    public static class Phone extends Item {
        public String text() {
            return "📱 A mysterious voice reveals insights from the future";
        }
        public void use(PrintWriter out, String player) {
            if (shells.size() <= 1) {
                System.out.println("How unfortunate...");
            } else {
                int selected = 1 + new Random().nextInt(shells.size() - 1);
                System.out.println("Shell #" + (selected + 1) + ", " + shells.get(selected));
            }
            sendMessage(out, "summary:", "Opponent used phone.\n");
        }
    }

    public static class Medicine extends Item {
        public String text() {
            return "💊 50% chance to gain 2 HP. If not, lose 1 HP.";
        }
        public void use(PrintWriter out, String player) {
            if (new Random().nextBoolean()) {
                System.out.println("You gained 2 HP!");
                sendMessage(out, "heal:", player + ",2,Opponent gained 2 HP!\n");
            } else {
                System.out.println("You collapsed! -1 HP");
                sendMessage(out, "heal:", player + ",-1,Opponent collapsed! They lose 1 HP.");
            }
        }
    }

    public static class Inverter extends Item {
        public String text() {
            return "🪫 Reverses polarity of current shell.";
        }
        public void use(PrintWriter out, String player) {
            if (!shells.isEmpty()) {
                if (shells.get(0) == Shell.live) {
                    shells.set(0, Shell.blank);
                } else{
                    shells.set(0, Shell.live);
                }
            }
            System.out.println("Inverted shell.");
            sendMessage(out, "invert:", "\n");
        }
    }
}
