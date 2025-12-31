Tor Scraper 

Bu proje, Tor ağı üzerindeki (.onion) ve clearnet web sitelerinden anonim olarak veri toplamak (HTML Kaynak Kodu + Ekran Görüntüsü) amacıyla Go dili ile geliştirilmiştir.

##  Özellikler

* **Tam Otomasyon:** Tor Browser'ı otomatik başlatır ve bağlantı kurulana kadar bekler.
* **Veri Toplama:** Hedef sitelerin ekran görüntüsünü (`.png`) ve HTML kodlarını (`.html`) kaydeder.
* **Gizlilik:** Tüm trafik SOCKS5 proxy üzerinden Tor ağına yönlendirilir.
* **Hata Analizi:** Erişilemeyen sitelerin nedenini  detaylıca raporlar.


1. Hedef Site Ekleme
Taranacak adresleri projenin ana dizinindeki targets.yaml dosyasına aşağıdaki formatta ekleyebilirsiniz:

sites:
  - [http://adresi-buraya-yazin.onion](http://adresi-buraya-yazin.onion)
  - [https://google.com](https://google.com)
  - [http://baska-bir-site.onion](http://baska-bir-site.onion)

  2. Uygulamayı Çalıştırma
Taramayı başlatmak için terminale şu komutu girin:

```go run main.go```