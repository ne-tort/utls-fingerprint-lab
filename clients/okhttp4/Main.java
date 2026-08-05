package okhttp4;

import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.Response;
import javax.net.ssl.*;
import java.security.cert.X509Certificate;
import java.util.concurrent.TimeUnit;

public class Main {
  public static void main(String[] args) throws Exception {
    TrustManager[] trustAll = new TrustManager[]{
      new X509TrustManager() {
        public void checkClientTrusted(X509Certificate[] c, String a) {}
        public void checkServerTrusted(X509Certificate[] c, String a) {}
        public X509Certificate[] getAcceptedIssuers() { return new X509Certificate[0]; }
      }
    };
    SSLContext ctx = SSLContext.getInstance("TLS");
    ctx.init(null, trustAll, new java.security.SecureRandom());
    OkHttpClient client = new OkHttpClient.Builder()
      .sslSocketFactory(ctx.getSocketFactory(), (X509TrustManager) trustAll[0])
      .hostnameVerifier((h, s) -> true)
      .connectTimeout(20, TimeUnit.SECONDS)
      .readTimeout(20, TimeUnit.SECONDS)
      .build();
    Request req = new Request.Builder()
      .url("https://okhttp4.fp.lab.local:8443/")
      .header("X-Target-Id", "okhttp4")
      .build();
    try (Response resp = client.newCall(req).execute()) {
      System.out.println(resp.code() + " " + resp.header("X-Captured-JA4") + " " + resp.body().string());
    }
  }
}
