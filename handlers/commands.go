package handlers

const (
    CommandStart            = "start"
    CommandHelp            = "help"
    CommandSearchStation   = "istasyonara"
    CommandSubscribe       = "abone"
    CommandListSubscriptions = "aboneliklerim"
    CancelSubscriptionPrefix = "cancel_subscription_"
)

type SubscriptionInfo struct {
    ID              int64
    DepartureStation string
    ArrivalStation   string
    TravelDate       string
}

var CommandDescriptions = map[string]string{
    CommandStart: "🚉 TCDD YHT Bilet Botu'na hoş geldiniz!\n\n" +
        "Bu bot ile neler yapabilirsiniz?\n\n" +
        "1️⃣ İstasyon Arama\n" +
        "   /istasyonara komutu ile istasyon adlarını öğrenebilirsiniz\n" +
        "   Örnek: /istasyonara ankara\n\n" +
        "2️⃣ Bilet Takibi\n" +
        "   /abone komutu ile bilet takibi başlatabilirsiniz\n" +
        "   Örnek: /abone ANKARA GAR-İSTANBUL(BOSTANCI)-29-12-2024\n\n" +
        "3️⃣ Takip Listesi\n" +
        "   /aboneliklerim komutu ile mevcut takiplerinizi yönetebilirsiniz\n\n" +
        "❓ Yardım için /help yazabilirsiniz",

    CommandHelp: "📋 *Detaylı Komut Rehberi*\n\n" +
        "*1. İstasyon Arama:*\n" +
        "Komut: /istasyonara <şehir adı>\n" +
        "• Büyük/küçük harf fark etmez\n" +
        "• Kısmi aramalar desteklenir\n" +
        "• İstediğiniz şehirdeki tüm istasyonları listeler\n" +
        "Örnek: /istasyonara ankara\n\n" +
        "*2. Bilet Takibi:*\n" +
        "Komut: /abone KALKIŞ-VARIŞ-GG-AA-YYYY\n" +
        "• İstasyon adlarını tam ve doğru girmelisiniz\n" +
        "• Tarih formatı GG-AA-YYYY şeklinde olmalıdır\n" +
        "• Her güzergah için bir takip açabilirsiniz\n" +
        "Örnek: /abone ANKARA GAR-İSTANBUL(BOSTANCI)-29-12-2024\n\n" +
        "*3. Takip Listesi:*\n" +
        "Komut: /aboneliklerim\n" +
        "• Aktif takiplerinizi görüntüler\n" +
        "• Tek tıkla takibi sonlandırabilirsiniz\n\n" +
        "*Önemli Bilgiler:*\n" +
        "• YHT bulunduğunda anında bildirim alırsınız\n" +
        "• Diğer trenler için saatlik bildirim gönderilir\n" +
        "• Geçmiş tarihli takipler otomatik silinir",

    CommandSearchStation: "🔍 *İstasyon Arama*\n\n" +
        "Hangi şehirdeki istasyonları aramak istiyorsunuz?\n" +
        "İstasyon adını öğrenmek için şehir adını yazın:\n\n" +
        "Örnek kullanımlar:\n" +
        "• /istasyonara ankara\n" +
        "• /istasyonara izmir\n" +
        "• /istasyonara ist\n\n" +
        "İpucu: Bilet takibi için istasyon adını buradan bulduğunuz şekilde kullanın.",

    CommandSubscribe: "🎫 *Bilet Takibi*\n\n" +
        "Takip etmek istediğiniz seferi şu formatta girin:\n\n" +
        "/abone KALKIŞ-VARIŞ-TARİH\n\n" +
        "Örnek:\n" +
        "/abone ANKARA GAR-İSTANBUL(BOSTANCI)-29-12-2024\n\n" +
        "Dikkat edilecek noktalar:\n" +
        "• İstasyon adları tam olmalı (önce /istasyonara ile kontrol edin)\n" +
        "• Tarih formatı: GG-AA-YYYY\n" +
        "• Tire (-) işaretini doğru kullanın\n\n" +
        "Bot sizin için düzenli olarak kontrol edecek ve bilet bulunduğunda haber verecektir.",

    CommandListSubscriptions: "📋 Aktif takip listesi\n\n" +
        "Kullanım: /aboneliklerim\n\n" +
        "Bu komut ile:\n" +
        "• Tüm aktif takiplerinizi görebilirsiniz\n" +
        "• İstemediğiniz takipleri durdurabilirsiniz\n" +
        "• Her bir takibin detaylarını görebilirsiniz",
}

const (
    // Kullanıcı mesajları
    MsgInvalidStationSearch = "🔎 Lütfen aramak istediğiniz istasyon için bir şehir adı yazın:\n\n" +
        "Örnek: /istasyonara ankara\n\n" +
        "İpucu: Kısmi aramalar da çalışır (örn: 'ist' yazarak İstanbul'daki istasyonları bulabilirsiniz)"

    MsgInvalidSubscription = "📝 Lütfen seferi şu şekilde girin:\n\n" +
        "/abone KALKIŞ-VARIŞ-TARİH\n\n" +
        "Örnek:\n" +
        "/abone ANKARA GAR-İSTANBUL(BOSTANCI)-29-12-2024\n\n" +
        "İpuçları:\n" +
        "• İstasyon adlarını /istasyonara ile kontrol edin\n" +
        "• Tarih formatı GG-AA-YYYY şeklinde olmalı\n" +
        "• Tire (-) işaretlerini unutmayın"
)
