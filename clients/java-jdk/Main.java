import javax.net.ssl.*;
import java.io.InputStream;
import java.net.*;
import java.security.cert.X509Certificate;

public class Main {
    public static void main(String[] args) throws Exception {
        String id = env("TARGET_ID", "java-jdk");
        String dial = env("DIAL_HOST", "capture:8443");
        String sni = id + ".fp.lab.local";
        String[] hp = dial.split(":");
        String host = hp[0];
        int port = Integer.parseInt(hp[1]);

        TrustManager[] trustAll = new TrustManager[]{
                new X509TrustManager() {
                    public void checkClientTrusted(X509Certificate[] c, String a) {}
                    public void checkServerTrusted(X509Certificate[] c, String a) {}
                    public X509Certificate[] getAcceptedIssuers() { return new X509Certificate[0]; }
                }
        };
        SSLContext ctx = SSLContext.getInstance("TLS");
        ctx.init(null, trustAll, new java.security.SecureRandom());

        Socket raw = new Socket(host, port);
        SSLSocket sock = (SSLSocket) ctx.getSocketFactory().createSocket(raw, sni, port, true);
        sock.startHandshake();

        String req = "GET / HTTP/1.1\r\nHost: " + sni + "\r\nX-Target-Id: " + id + "\r\nConnection: close\r\n\r\n";
        sock.getOutputStream().write(req.getBytes());
        InputStream in = sock.getInputStream();
        byte[] buf = new byte[512];
        int n = in.read(buf);
        if (n > 0) System.out.write(buf, 0, n);
        sock.close();
    }

    static String env(String k, String d) {
        String v = System.getenv(k);
        return (v == null || v.isEmpty()) ? d : v;
    }
}
