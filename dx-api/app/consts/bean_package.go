package consts

// BeanPackage defines a purchasable bean package.
type BeanPackage struct {
	Price int // price in fen (1 yuan = 100 fen)
	Beans int // base bean amount
	Bonus int // bonus bean amount
}

// BeanPackages maps package slugs to their definitions.
var BeanPackages = map[string]BeanPackage{
	"beans-1":   {Price: 100, Beans: 1000, Bonus: 0},
	"beans-5":   {Price: 500, Beans: 5000, Bonus: 0},
	"beans-10":  {Price: 1000, Beans: 10000, Bonus: 1000},
	"beans-50":  {Price: 5000, Beans: 50000, Bonus: 7500},
	"beans-100": {Price: 10000, Beans: 100000, Bonus: 20000},
}

// BeanPackageSlugs lists valid package slugs for validation.
var BeanPackageSlugs = []string{"beans-1", "beans-5", "beans-10", "beans-50", "beans-100"}
