package handlers

const (
	CommandStart             = "start"
	CommandHelp              = "help"
	CommandSearchStation     = "istasyonara"
	CommandSubscribe         = "abone"
	CommandListSubscriptions = "aboneliklerim"
	CancelSubscriptionPrefix = "cancel_subscription_"
)

type SubscriptionInfo struct {
	ID               int64
	DepartureStation string
	ArrivalStation   string
	TravelDate       string
}

var CommandDescriptions = map[string]string{
	CommandStart: "🚉 *TCDD YHT Bilet Botu'na hoş geldiniz!*\n\n" +
		"*Bu bot ile neler yapabilirsiniz?*\n\n" +
		"🎫 *Bilet Takibi*\n" +
		"   • /abone komutu ile bilet takibi başlatın\n" +
		"   • YHT bulunduğunda anında bildirim alın\n\n" +
		"📋 *Takip Listesi*\n" +
		"   • /aboneliklerim ile takiplerinizi yönetin\n" +
		"   • Tek tıkla takibi sonlandırın\n\n" +
		"❓ Detaylı bilgi için /help yazabilirsiniz",

	CommandHelp: "📋 *Detaylı Komut Rehberi*\n\n" +
		"*1. Bilet Takibi* (/abone)\n" +
		"   • İstasyon adı yazarak arama yapın\n" +
		"   • Kalkış ve varış istasyonlarını seçin\n" +
		"   • Tarih seçimini kolayca yapın\n" +
		"   • YHT bulunduğunda anında haberdar olun\n\n" +
		"*2. Takip Listesi* (/aboneliklerim)\n" +
		"   • Tüm aktif takiplerinizi görüntüleyin\n" +
		"   • İstemediğiniz takibi tek tıkla durdurun\n\n" +
		"*Önemli Bilgiler:*\n" +
		"   • YHT bulunduğunda anında bildirim 🔔\n" +
		"   • Diğer trenler için saatlik kontrol ⏰\n" +
		"   • Otomatik geçmiş takip temizleme 🧹",

	CommandSearchStation: "🔍 *İstasyon Arama*\n\n" +
		"*İstasyon adını öğrenmek için şehir adını yazın:*\n\n" +
		"*Örnek Kullanımlar:*\n" +
		"   • /istasyonara ankara\n" +
		"   • /istasyonara izmir\n" +
		"   • /istasyonara ist\n\n" +
		"💡 *İpucu:* Kısmi aramalar da çalışır\n" +
		"Örn: 'ist' yazarak İstanbul'daki tüm istasyonları bulabilirsiniz",

	CommandSubscribe: "🎫 *Bilet Takibi Başlatılıyor*\n\n" +
		"Adım 1️⃣: Kalkış istasyonunu seçin\n" +
		"• İstasyon adını yazın (örn: ankara)\n" +
		"• Listeden seçim yapın\n\n" +
		"💡 *İpucu:* En az 2 karakter girmelisiniz",

	CommandListSubscriptions: "📋 *Aktif Takip Listesi*\n\n" +
		"*Bu komut ile:*\n" +
		"• 📍 Tüm aktif takiplerinizi görüntüleyin\n" +
		"• ❌ İstemediğiniz takipleri durdurun\n" +
		"• 🕒 Takip detaylarını kontrol edin\n\n" +
		"💡 Takipleriniz otomatik olarak güncel tutulur",
}

const (
	// Kullanıcı mesajları
	MsgInvalidStationSearch = "🔎 *İstasyon Arama*\n\n" +
		"*Nasıl Kullanılır?*\n" +
		"• /istasyonara ŞEHIR_ADI şeklinde arama yapın\n\n" +
		"*Örnek:*\n" +
		"• /istasyonara ankara\n\n" +
		"💡 *İpucu:* Kısmi kelimeler de çalışır\n" +
		"Örn: 'ist' yazarak İstanbul'daki istasyonları bulabilirsiniz"

	MsgInvalidSubscription = "📝 *Abonelik Formatı*\n\n" +
		"*Doğru Format:*\n" +
		"/abone KALKIŞ-VARIŞ-TARİH\n\n" +
		"*Örnek:*\n" +
		"/abone ANKARA GAR-İSTANBUL(BOSTANCI)-29-12-2024\n\n" +
		"*İpuçları:*\n" +
		"• 🔍 İstasyon adları için /istasyonara kullanın\n" +
		"• 📅 Tarih formatı: GG-AA-YYYY\n" +
		"• ℹ️ Tire (-) işaretlerini unutmayın"
)
