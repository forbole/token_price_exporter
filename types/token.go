package token

type Config struct {
	Tokens []Token `yaml:"tokens"`
}

type Token struct {
	Denom string `yaml:"denom"`
	ID    string `yaml:"id"`
}
