package src;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.io.PrintWriter;
import java.net.ServerSocket;
import java.net.Socket;
import java.util.*;
import java.util.concurrent.CopyOnWriteArrayList;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.atomic.AtomicBoolean;


public class BuckshotRoulette {
    // Constants
    private static final int PORT = 0xDEA;
    private static final int BUFFER_SIZE = 1024;

    public enum Shell {
        live(0),
        blank(1);

        private final int value;
        Shell(int v) { value = v; }
        public int getValue() { return value; }
        public static Shell fromValue(int v) {
            return (v == 0) ? live : blank;
        }
    }

    static Map<String,Integer> hp = new HashMap<>();
    static List<Shell> shells = new CopyOnWriteArrayList<>();
    static int damage = 1;
    static final Map<Item, Integer> items = new HashMap<>();
    static boolean cuffedOpponent = false;
    private static final AtomicBoolean gameOver = new AtomicBoolean(false);
    private static final Map<String, Item> numberToItem = new HashMap<>();

    private static final Object mutex = new Object();

    static {
        numberToItem.put("3", new Item.MagnifyingGlass());
        numberToItem.put("4", new Item.Cigarette());
        numberToItem.put("5", new Item.Beer());
        numberToItem.put("6", new Item.Handsaw());
        numberToItem.put("7", new Item.Handcuffs());
        numberToItem.put("8", new Item.Phone());
        numberToItem.put("9", new Item.Medicine());
        numberToItem.put("10", new Item.Inverter());

        for (Item item : numberToItem.values()) {
            items.put(item, 0);
        }
    }

    private static void moreItems(PrintWriter out) {
        List<Item> tempItems = new ArrayList<>(numberToItem.values());
        Random random = new Random();
        List<Item> chosenItems = new ArrayList<>();
        for (int i = 0; i < 2; i++) {
            Item selectedItem = tempItems.get(random.nextInt(tempItems.size()));
            chosenItems.add(selectedItem);
            items.put(selectedItem, items.get(selectedItem) + 1);
        }
        StringBuilder stringBuilder = new StringBuilder();
        for (Item item : chosenItems) {
            stringBuilder.append(item.getClass().getName()).append(", ");
        }
        stringBuilder.append("\n");
        System.out.println("You get " + stringBuilder);
        sendMessage(out, "summary:", "Opponent gets " + stringBuilder + "\n");
    }

    private static void displayHealth(String player, int hp) {
        System.out.println(player + "'s HP: " + hp + "\n");
    }

    static void sendMessage(PrintWriter out, String prefix, String msg) {
        out.println(prefix + msg);
        out.flush();
    }

    private static List<Shell> loadShotgun(PrintWriter out) {
        Random rand = new Random();
        int live_shells = rand.nextInt(4) + 1;
        int blank_shells = rand.nextInt(4) + 1;

        List<Shell> newShells = new ArrayList<>();
        for (int i = 0; i < live_shells; i++) {
            newShells.add(Shell.live);
        }
        for (int i = 0; i < blank_shells; i++) {
            newShells.add(Shell.blank);
        }
        Collections.shuffle(newShells);

        System.out.println("[INFO] Shotgun loaded with " + live_shells + " live shells and "
                + blank_shells + " blank shells (order is hidden).\n");

        StringBuilder sb = new StringBuilder();
        for (Shell s : newShells) {
            sb.append(s.getValue());
        }
        sb.append("\n");
        sendMessage(out, "reload:", sb.toString());
        moreItems(out);

        return newShells;
    }

    private static String printRoundSummary(String shooter, String target, Shell shell) {
        String outcome = (shell == Shell.live) ? "Hit (Live shell)" : "Miss (Blank shell)";
        String shot_description = shooter.equals(target) ? "themselves" : target;

        StringBuilder summary = new StringBuilder();
        summary.append("\n--- Round Summary ---\n");
        summary.append("Move: ").append(shooter).append(" shot ").append(shot_description).append(" with a ")
                .append((shell == Shell.live) ? "live" : "blank").append(" shell.\n");
        summary.append("Outcome: ").append(outcome).append("\n");
        summary.append("Remaining HP:\n");
        for (Map.Entry<String,Integer> entry : hp.entrySet()) {
            summary.append(entry.getKey()).append(": ").append(entry.getValue()).append(" HP\n");
        }
        summary.append("---------------------\n\n");
        return summary.toString();
    }

    private static Triple<List<Shell>, Map<String,Integer>, String> takeTurn(
            String target, String other, PrintWriter out,
            String shooter)
    {
        if (shells.isEmpty()) {
            // Should not happen if logic is correct
            return new Triple<>(shells, hp, "switch");
        }

        Shell shell = shells.remove(0);
        System.out.println(shooter + " pulls the trigger... it's a " + shell.name() + " shell!");
        sendMessage(out, "action:", shooter + " fired a " + shell.name() + " shell at " + target + "!");

        String action;
        if (shell == Shell.live) {
            hp.put(target, hp.get(target) - damage);
            displayHealth(target, hp.get(target));
            sendMessage(out, "summary:", target + " lost " + damage + " HP. Remaining HP: " + hp.get(target));
            sendMessage(out, "damage:", damage + "," + target);
            damage = 1;

            if (hp.get(target) <= 0) {
                System.out.println(target + " has been eliminated! " + other + " wins!");
                sendMessage(out, "control:", "game_over");
                System.exit(0);
            }
            if (cuffedOpponent) {
                cuffedOpponent = false;
                System.out.println("Your opponent is cuffed!");
                action = "continue";
            } else {
                action = "switch";
            }
        } else {
            // Blank shell
            if (shooter.equals(target)) {
                action = "continue";
            } else {
                if (cuffedOpponent) {
                    cuffedOpponent = false;
                    System.out.println("Your opponent is cuffed!");
                    action = "continue";
                } else {
                    action = "switch";
                }
            }
        }

        String summary = printRoundSummary(shooter, target, shell);
        System.out.println(summary);
        sendMessage(out, "summary:", summary);

        return new Triple<>(shells, hp, action);
    }

    private static void handleIncomingMessages(BufferedReader in, PrintWriter out,
                                               String opponent,
                                               String player_name) {
        StringBuilder buffer = new StringBuilder();
        try {
            char[] buff = new char[BUFFER_SIZE];
            while (!gameOver.get()) {
                int read = in.read(buff);
                if (read == -1) {
                    System.out.println("Connection lost. Exiting...");
                    System.exit(1);
                }
                buffer.append(buff, 0, read);

                int newlineIndex;
                synchronized (mutex) {
                    while ((newlineIndex = buffer.indexOf("\n")) != -1) {
                        String line = buffer.substring(0, newlineIndex).trim();
                        buffer.delete(0, newlineIndex + 1);
                        if (line.isEmpty()) {
                            continue;
                        }

                        if (line.startsWith("control:")) {
                            String msg = line.substring(8);
                            switch (msg) {
                                case "game_over" -> {
                                    System.out.println("Game Over! Exiting...");
                                    System.exit(0);
                                }
                                case "continue" ->
                                        System.out.println(opponent + " got a blank! It's still their turn.");
                                case "switch" -> {
                                    // Opponent ended their turn
                                }
                                case "your_turn" ->
                                    // It's now our turn
                                        currentTurn(player_name, opponent, out);
                                default -> System.out.println("[UNKNOWN CONTROL MESSAGE]: " + msg);
                            }
                        } else if (line.startsWith("summary:")) {
                            String summary_msg = line.substring(8);
                            System.out.print(summary_msg);
                        } else if (line.startsWith("action:")) {
                            // Remove a shell from our local vector if any
                            if (!shells.isEmpty()) {
                                shells.remove(0);
                            }
                            String action_msg = line.substring(7);
                            System.out.println(opponent + "'s move: " + action_msg);
                        } else if (line.startsWith("reload:")) {
                            shells.clear();
                            String reload_msg = line.substring(7);
                            int live_count = 0;
                            int blank_count = 0;
                            for (char c : reload_msg.toCharArray()) {
                                Shell s = Shell.fromValue(Character.getNumericValue(c));
                                shells.add(s);
                                if (s == Shell.live) {
                                    live_count++;
                                } else {
                                    blank_count++;
                                }
                            }
                            System.out.println("[INFO] Shotgun loaded with " + live_count + " live shells and "
                                    + blank_count + " blank shells (order is hidden).\n");
                        } else if (line.startsWith("damage:")) {
                            String info_msg = line.substring("damage:".length());
                            System.out.println(info_msg);
                            String[] parts = info_msg.split(",");
                            String target_player = parts[1];
                            int new_hp = Integer.parseInt(parts[0]);
                            hp.put(target_player, hp.get(target_player) - new_hp);
                        } else if (line.startsWith("moreitems:")) {
                            moreItems(out);
                        } else if (line.startsWith("eject:")) {
                            if (!shells.isEmpty()) {
                                shells.remove(0);
                            }
                            System.out.println(line.substring("eject:".length()));
                        } else if (line.startsWith("heal:")) {
                            String substring = line.substring("heal:".length()).trim();
                            String[] params = substring.split(",");
                            hp.put(params[0], hp.get(params[0]) + Integer.parseInt(params[1]));
                            System.out.println(params[2]);
                        } else if (line.startsWith("invert:")) {
                            if (!shells.isEmpty()) {
                                if (shells.get(0) == Shell.live) {
                                    shells.set(0, Shell.blank);
                                } else{
                                    shells.set(0, Shell.live);
                                }
                            }
                            System.out.println("Opponent inverted shell.");
                        }
                        else {
                            // Unrecognized message
                            System.out.println(line);
                        }
                    }
                }
            }
        } catch (IOException e) {
            System.out.println("Connection lost. Exiting...");
            System.exit(1);
        }
    }

    private static void currentTurn(String player, String opponent, PrintWriter out) {
        Scanner sc = new Scanner(System.in);
        while (!gameOver.get()) {
            if (shells.isEmpty()) {
                sendMessage(out, "moreitems:", "\n");
                System.out.println("Reloading the shotgun!");
                shells = loadShotgun(out);
            }
            synchronized (mutex) {
                System.out.println("Options:");
                System.out.println("1. Shoot Yourself");
                System.out.println("2. Shoot Your Opponent");
                for (Map.Entry<String, Item> entry : numberToItem.entrySet()) {
                    System.out.println(entry.getKey() + ". " + entry.getValue().text() + " (" + items.get(entry.getValue()) + ")");
                }
                System.out.print("Choose an option: ");
            }
            String choice = sc.nextLine().trim();

            if (choice.equals("1") || choice.equals("2")) {
                try {
                    Triple<List<Shell>, Map<String,Integer>, String> result;
                    if (choice.equals("1")) {
                        result = takeTurn(player, opponent, out, player);
                    } else {
                        result = takeTurn(opponent, player, out, player);
                    }

                    shells = result.a;
                    hp = result.b;
                    String action = result.c;

                    if (action.equals("switch")) {
                        sendMessage(out, "control:", "your_turn");
                        System.out.println("\nTurn passes to " + opponent + ".\n");
                        break;
                    } else if (action.equals("continue")) {
                        System.out.println("\n" + player + " gets another turn!\n");
                    }
                } catch (Exception e) {
                    System.out.println("Connection lost. Exiting...");
                    System.exit(1);
                }
            } else if (numberToItem.containsKey(choice) && items.get(numberToItem.get(choice)) > 0) {
                numberToItem.get(choice).use(out, player);
                items.put(numberToItem.get(choice), items.get(numberToItem.get(choice)) - 1);
            } else if (numberToItem.containsKey(choice)) {
                System.out.println("You don't have " + numberToItem.get(choice).getClass().getName() + "!");
            }
            else {
                System.out.println("Invalid choice.");
            }
        }
    }

    public static void main(String[] args) {
        System.out.println("Welcome to P2P Buckshot Roulette!\n");
        Scanner sc = new Scanner(System.in);

        System.out.print("Enter your name: ");
        String player_name = sc.nextLine().trim();

        String mode;
        do {
            System.out.print("Do you want to host the game or join? (host/join): ");
            mode = sc.nextLine().trim().toLowerCase();
        } while (!mode.equals("host") && !mode.equals("join"));

        Socket connection;
        String opponent_name;
        BufferedReader in;
        PrintWriter out;

        CountDownLatch turnLatch = new CountDownLatch(1);

        try {
            if (mode.equals("host")) {
                ServerSocket serverSocket = new ServerSocket(PORT);
                System.out.println("Waiting for a connection on port " + PORT + "...");

                BroadcastThread broadcaster = new BroadcastThread(player_name, PORT);
                broadcaster.start();

                connection = serverSocket.accept();
                System.out.println("Connection established with " + connection.getRemoteSocketAddress() + "\n");

                broadcaster.shutdownBroadcast();

                out = new PrintWriter(connection.getOutputStream(), true);
                in = new BufferedReader(new InputStreamReader(connection.getInputStream()));

                // Send our name
                out.print(player_name + "\n");
                out.flush();

                char[] buf = new char[BUFFER_SIZE];
                int len = in.read(buf);
                if (len <= 0) {
                    System.out.println("Failed to receive opponent name.");
                    System.exit(1);
                }
                opponent_name = new String(buf, 0, len).trim();

                serverSocket.close();
            } else {
                DiscoveryThread discovery = new DiscoveryThread();
                discovery.start();

                System.out.println("[INFO] Searching for a src.BuckshotRoulette game on LAN...");
                System.out.println("       (This might take a few seconds; press Enter to cancel.)");

                // Optionally, you could allow the user to press Enter to cancel the search
                // or wait a maximum time. For simplicity, let's wait up to ~10 seconds (the
                // socket timeout in src.DiscoveryThread).
                sc.nextLine();
                discovery.stopDiscovery();
                discovery.join();

                String discoveredHost = discovery.getDiscoveredHost();
                int discoveredPort = discovery.getDiscoveredPort();

                if (discoveredHost == null || discoveredPort < 0) {
                    System.out.println("[ERROR] No LAN game found. Exiting...");
                    System.exit(0);
                }

                connection = new Socket(discoveredHost, PORT);
                System.out.println("Connected to the host!\n");

                out = new PrintWriter(connection.getOutputStream(), true);
                in = new BufferedReader(new InputStreamReader(connection.getInputStream()));

                char[] buf = new char[BUFFER_SIZE];
                int len = in.read(buf);
                if (len <= 0) {
                    System.out.println("Failed to receive host name.");
                    System.exit(1);
                }
                opponent_name = new String(buf, 0, len).trim();

                while (player_name.equals(opponent_name)) {
                    System.out.print("Enter another name - name cannot be same as host: ");
                    player_name = sc.nextLine().trim();
                }

                out.print(player_name + "\n");
                out.flush();
            }

            System.out.println("""
                    How to play:
                    Get your opponent's HP to 0.
                    The shotgun is loaded with a random number of blank and live shells in an unknown order.
                    If you shoot the opponent with a live shell, they lose HP and turn passes over.
                    If you shoot yourself with a live shell, you lose HP and turn passes over.
                    If you shoot your opponent with a blank, turn passes to the opponent.
                    If you shoot yourself with a blank, you get another turn.
                    Shotgun reloads when it becomes empty.
                    
                    Two items each. More items after each reload.
                    -----------""");

            hp.put(player_name, 5);
            hp.put(opponent_name, 5);

            displayHealth(player_name, hp.get(player_name));
            displayHealth(opponent_name, hp.get(opponent_name));

            // Start listening for incoming messages in a separate thread
            PrintWriter outFinal = out;
            BufferedReader inFinal = in;
            String opponentFinal = opponent_name;

            String finalPlayer_name = player_name;
            Thread incomingThread = new Thread(() -> handleIncomingMessages(inFinal, outFinal, opponentFinal, finalPlayer_name));
            incomingThread.setDaemon(true);
            incomingThread.start();

            if (mode.equals("host")) {
                currentTurn(player_name, opponent_name, out);
                System.out.println("Waiting for your opponent's turn...");
                turnLatch.await();
            } else {
                System.out.println("Waiting for your turn...");
                turnLatch.await();
            }

        } catch (Exception e) {
            e.printStackTrace();
            System.exit(1);
        }
    }

    // A simple triple container
    private static class Triple<A,B,C> {
        A a;
        B b;
        C c;
        Triple(A a, B b, C c) {
            this.a = a;
            this.b = b;
            this.c = c;
        }
    }
}
