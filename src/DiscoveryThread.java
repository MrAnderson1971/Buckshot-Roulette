package src;

import java.net.DatagramPacket;
import java.net.DatagramSocket;
import java.net.InetAddress;

class DiscoveryThread extends Thread {
    private volatile boolean running = true;

    // We store the discovered game’s IP/hostname and port
    private String discoveredHost = null;
    private int discoveredPort = -1;

    @Override
    public void run() {
        try (DatagramSocket socket = new DatagramSocket(BroadcastThread.BROADCAST_PORT, InetAddress.getByName("0.0.0.0"))) {
            socket.setSoTimeout(10000); // 10-second timeout or adjust as needed

            while (running) {
                byte[] buffer = new byte[1024];
                DatagramPacket packet = new DatagramPacket(buffer, buffer.length);

                try {
                    // Wait for a broadcast
                    socket.receive(packet);
                    String msg = new String(packet.getData(), 0, packet.getLength()).trim();

                    if (msg.startsWith("BUCKSHOT_ROULETTE:")) {
                        // Expected format: BUCKSHOT_ROULETTE:<name>:<port>
                        String[] parts = msg.split(":");
                        if (parts.length == 3) {
                            discoveredHost = packet.getAddress().getHostAddress();
                            discoveredPort = Integer.parseInt(parts[2]);

                            System.out.println("[INFO] Found a src.BuckshotRoulette game!");
                            System.out.println("       Host IP: " + discoveredHost);
                            System.out.println("       Port   : " + discoveredPort);

                            // We can stop after the first discovery or keep going
                            running = false;
                        }
                    }
                } catch (Exception ex) {
                    // If we time out or other exceptions, just keep waiting or break
                    System.out.println("[INFO] src.DiscoveryThread: still searching for a host...");
                }
            }
        } catch (Exception e) {
            System.err.println("[ERROR] src.DiscoveryThread: " + e.getMessage());
        }
    }

    public void stopDiscovery() {
        running = false;
    }

    public String getDiscoveredHost() {
        return discoveredHost;
    }

    public int getDiscoveredPort() {
        return discoveredPort;
    }
}
