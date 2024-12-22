package src;

import java.io.PrintWriter;

import static src.BuckshotRoulette.*;

public abstract class Item {
    public abstract String text();
    public abstract void use(PrintWriter out, String player);

    public static class MagnifyingGlass extends Item {
        public String text() {
            return "🔍 Reveals the next shell.";
        }
        public void use(PrintWriter out, String player) {
            System.out.println("The next item is a " + shells.getFirst().toString() + " shell.\n");
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
            sendMessage(out, "summary:", player + "\n");
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
            Shell first = shells.getFirst();
            shells.removeFirst();
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
}
