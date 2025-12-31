package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"golang.org/x/net/proxy"
	"gopkg.in/yaml.v3"
)

const (
	TorProxyAddress = "127.0.0.1:9150"
	InputFile       = "targets.yaml"
	OutputDir       = "data"
	ReportFile      = "scan_report.log"
	UserAgentStr    = "Mozilla/5.0 (Windows NT 10.0; rv:109.0) Gecko/20100101 Firefox/115.0"
	TorExePath      = `C:\Users\dgnmu\OneDrive\Masaüstü\Tor Browser\Browser\firefox.exe`
)

type Config struct {
	Sites []string `yaml:"sites"`
}

func main() {
	fmt.Println("#################################################")
	fmt.Println("#           TOR SCRAPER'A HOSGELDİNİZ           #")
	fmt.Println("#################################################")

	prepareEnvironment()

	targets, err := readTargets(InputFile)
	if err != nil {
		log.Fatalf("[HATA] Hedef listesi okunamadı: %v", err)
	}

	TorIsRunning()

	if !checkTorConnection() {
		log.Fatal("[HATA] Tor ağına çıkış yapılamadı!")
	}

	fmt.Println("\n[BILGI] Tarayıcı hazırlanıyor...")

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ProxyServer("socks5://"+TorProxyAddress),
		chromedp.UserAgent(UserAgentStr),
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.WindowSize(1366, 768),
	)

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()

	browserCtx, cancelBrowser := chromedp.NewContext(allocCtx)
	defer cancelBrowser()

	if err := chromedp.Run(browserCtx); err != nil {
		log.Fatalf("[HATA] Tarayıcı başlatılamadı: %v", err)
	}

	fmt.Println("--------------------- Tarama Süreci -----------------------")

	for _, url := range targets {
		if url == "" {
			continue
		}

		fmt.Printf("[SCAN] %s ... ", url)

		tabCtx, cancelTab := chromedp.NewContext(browserCtx)
		timeoutCtx, cancelTimeout := context.WithTimeout(tabCtx, 60*time.Second)

		var screenshotBuf []byte
		var htmlContent string

		err := chromedp.Run(timeoutCtx,
			chromedp.Navigate(url),
			chromedp.Sleep(5*time.Second),
			chromedp.OuterHTML("html", &htmlContent),
			chromedp.FullScreenshot(&screenshotBuf, 90),
		)

		if err != nil {
			status, detail := classifyError(err)
			fmt.Printf("-> [%s] (%s)\n", status, detail)
			logToReport(url, status, detail)
		} else {
			saveScreenshot(url, screenshotBuf)
			saveHTML(url, htmlContent)
			fmt.Println("-> [BAŞARILI] HTML ve Screenshot kaydedildi.")
			logToReport(url, "BAŞARILI", "HTML ve Screenshot kaydedildi.")
		}

		cancelTimeout()
		cancelTab()
	}

	fmt.Println("\n[BITTI] Tarama bitti data klasörünü kontrol et.")
}

func checkTorConnection() bool {
	maxRetries := 10

	fmt.Println("\n[BILGI] Tor Ağ Bağlantısı Test Ediliyor ...")

	for i := 1; i <= maxRetries; i++ {

		fmt.Printf("   > Deneme %d/%d: ", i, maxRetries)

		dialer, err := proxy.SOCKS5("tcp", TorProxyAddress, nil, proxy.Direct)
		if err != nil {
			fmt.Printf("Proxy Hatası (%v) - Bekleniyor...\n", err)
			time.Sleep(5 * time.Second)
			continue
		}

		transport := &http.Transport{Dial: dialer.Dial}
		client := &http.Client{Transport: transport, Timeout: 10 * time.Second}

		req, _ := http.NewRequest("GET", "https://check.torproject.org/api/ip", nil)
		req.Header.Set("User-Agent", UserAgentStr)

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Bağlantı Yok - 5sn Bekleniyor\n")
			time.Sleep(5 * time.Second)
			continue
		}

		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("[BAŞARILI] Tor Bağlantısı Hazır! IP: %s\n", strings.TrimSpace(string(body)))
		return true
	}

	return false
}

func TorIsRunning() {
	fmt.Println("[BILGI] Tor servisi 9150 portu kontrol ediliyor...")

	conn, err := net.DialTimeout("tcp", TorProxyAddress, 1*time.Second)
	if err == nil {
		conn.Close()
		fmt.Println("[BAŞARILI] Tor portu açık. Ağ kontrolüne geçiliyor.")
		return
	}

	fmt.Println("[UYARI] Tor kapalı. Otomatik başlatılıyor...")
	cmd := exec.Command(TorExePath)
	err = cmd.Start()
	if err != nil {
		log.Fatalf("[HATA] Tor başlatılamadı! : %v", err)
	}

	fmt.Println("\n!!! Tor Browser açıldı.")
	fmt.Println("OTOMATİK BAĞLANTI OLUŞMUYORSA LÜTFEN CONNECT BUTONUNA BASINIZ !!!")

	fmt.Print("Uygulamanın açılması bekleniyor")
	for i := 0; i < 30; i++ {
		time.Sleep(2 * time.Second)
		fmt.Print(".")
		conn, err := net.DialTimeout("tcp", TorProxyAddress, 1*time.Second)
		if err == nil {
			conn.Close()
			fmt.Println("\n[BAŞARILI] Tor uygulaması açıldı.")
			return
		}
	}
	log.Fatal("\n[HATA] Tor uygulaması başlatılamadı.")
}

func classifyError(err error) (string, string) {
	msg := err.Error()

	if err == context.DeadlineExceeded {
		return "TIMEOUT", "Süre doldu (60sn)"
	}
	if err == context.Canceled {
		return "CRASH", "Tarayıcı kapandı"
	}

	if strings.Contains(msg, "ERR_SOCKS_CONNECTION_FAILED") {
		return "ONION_DEAD", "Onion sitesi bulunamadı veya çevrimdışı"
	}

	if strings.Contains(msg, "ERR_CONNECTION_REFUSED") {
		return "REFUSED", "Bağlantı reddedildi (Port kapalı)"
	}
	if strings.Contains(msg, "ERR_NAME_NOT_RESOLVED") {
		return "INVALID", "Adres çözülemedi (Link hatalı)"
	}
	if strings.Contains(msg, "ERR_TIMED_OUT") || strings.Contains(msg, "ERR_CONNECTION_TIMED_OUT") {
		return "TIMEOUT", "Ağ zaman aşımı (Sunucu yanıt vermedi)"
	}
	if strings.Contains(msg, "ERR_PROXY_CONNECTION_FAILED") {
		return "PROXY_ERR", "Tor proxy bağlantısı koptu"
	}

	if strings.Contains(msg, "0 width") || strings.Contains(msg, "-32000") {
		return "EMPTY_PAGE", "Sayfa yüklenemediği için görüntü alınamadı"
	}

	if strings.Contains(msg, "404") {
		return "NOT_FOUND", "Sayfa bulunamadı (404)"
	}

	// Bilinmeyen hata
	return "ERROR", msg
}

func readTargets(filename string) ([]string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("YAML parse hatası: %v", err)
	}
	return config.Sites, nil
}

func prepareEnvironment() {
	if _, err := os.Stat(OutputDir); os.IsNotExist(err) {
		os.Mkdir(OutputDir, 0755)
	}
	f, _ := os.Create(ReportFile)
	f.WriteString(fmt.Sprintf("Scan Report - %s\n-----------------------------------\n", time.Now().Format(time.RFC3339)))
	f.Close()
}

func logToReport(url, status, message string) {
	f, err := os.OpenFile(ReportFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	f.WriteString(fmt.Sprintf("[%s] %-10s | %s | Msg: %s\n", time.Now().Format("15:04:05"), status, url, message))
}

func saveScreenshot(url string, data []byte) {
	safeName := sanitizeFilename(url)
	path := filepath.Join(OutputDir, safeName+".png")
	os.WriteFile(path, data, 0644)
}

func saveHTML(url string, content string) {
	safeName := sanitizeFilename(url)
	path := filepath.Join(OutputDir, safeName+".html")
	os.WriteFile(path, []byte(content), 0644)
}

func sanitizeFilename(url string) string {
	safeName := strings.ReplaceAll(url, "http://", "")
	safeName = strings.ReplaceAll(safeName, "https://", "")
	safeName = strings.ReplaceAll(safeName, "/", "_")
	safeName = strings.ReplaceAll(safeName, ":", "")
	safeName = strings.ReplaceAll(safeName, ".onion", "")
	if len(safeName) > 50 {
		safeName = safeName[:50]
	}
	return safeName
}
