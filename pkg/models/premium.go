package models

const (
	OrderTypeCourse  = "course"
	OrderTypePremium = "premium"
)

// Premium paketlari (T-11, T-12). Fan (subject) shulardan biriga tegishli bo'ladi.
const (
	PackageFundamental = "fundamental"
	PackagePreclinical = "preclinical"
	PackageClinical    = "clinical"
	// PackageAll paketlarga bo'linishdan oldingi umumiy premium, faqat eski obunalarda uchraydi.
	PackageAll = "all"
)

var PremiumPackages = []string{PackageFundamental, PackagePreclinical, PackageClinical}

func IsPremiumPackage(p string) bool {
	for _, v := range PremiumPackages {
		if v == p {
			return true
		}
	}
	return false
}

type PremiumSettings struct {
	Package        string `json:"package"`
	Price          int    `json:"price"`
	DurationMonths int    `json:"duration_months"`
}

type PremiumOrderBody struct {
	Package     string `json:"package"`
	PromoCodeId string `json:"promocode_id"`
}

type PremiumOrderReq struct {
	UserId      string
	Package     string
	PromoCodeId string
}

type PackageStatus struct {
	Package  string  `json:"package"`
	IsActive bool    `json:"is_active"`
	Until    *string `json:"until"`
}

// AccessStatus foydalanuvchining premium va ustoz holati (T-12, T-15, T-16).
// is_premium va premium_until istalgan paket bo'yicha; paketma-paket holat packages ichida.
type AccessStatus struct {
	IsPremium    bool            `json:"is_premium"`
	PremiumUntil *string         `json:"premium_until"`
	IsUstoz      bool            `json:"is_ustoz"`
	Packages     []PackageStatus `json:"packages"`
}

type PremiumPackageInfo struct {
	Package        string  `json:"package"`
	Price          int     `json:"price"`
	DurationMonths int     `json:"duration_months"`
	IsActive       bool    `json:"is_active"`
	Until          *string `json:"until"`
}

type PremiumInfoRes struct {
	IsPremium    bool                 `json:"is_premium"`
	PremiumUntil *string              `json:"premium_until"`
	IsUstoz      bool                 `json:"is_ustoz"`
	Packages     []PremiumPackageInfo `json:"packages"`
}

type UstozBody struct {
	IsUstoz bool `json:"is_ustoz"`
}

type UstozReq struct {
	UserId  string
	IsUstoz bool
}
