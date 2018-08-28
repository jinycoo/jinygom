package cache

var mongoCfg *mongoConfig

type mongoConfig struct {
	Name string  `yaml:"name"`
}

func mongoConn(mf string) {

}