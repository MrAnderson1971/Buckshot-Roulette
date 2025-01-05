package src;

import java.net.DatagramPacket;
import java.net.DatagramSocket;
import java.net.InetAddress;

class BroadcastThread extends Thread {
    public static final int BROADCAST_PORT = 0x60D;
    private final String serverName;
    private final int gamePort;
    private volatile boolean running = true;

    public BroadcastThread(String serverName, int gamePort) {
        this.serverName = serverName;
        this.gamePort = gamePort;
    }

    @Override
    public void run() {
        try (DatagramSocket socket = new DatagramSocket()) {
            socket.setBroadcast(true);

            while (running) {
                String broadcastMessage = "BUCKSHOT_ROULETTE:" + serverName + ":" + gamePort;
                byte[] buffer = broadcastMessage.getBytes();

                // 255.255.255.255 is a limited-broadcast address (works in most LANs)
                DatagramPacket packet = new DatagramPacket(
                        buffer,
                        buffer.length,
                        InetAddress.getByName("255.255.255.255"),
                        BROADCAST_PORT
                );

                socket.send(packet);

                // Broadcast every 2 seconds (adjust as you like)
                Thread.sleep(2000);
            }
        } catch (Exception e) {
            System.err.println("[ERROR] src.BroadcastThread: " + e.getMessage());
        }
    }

    public void shutdownBroadcast() {
        running = false;
    }
}
