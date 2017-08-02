package conf

import (
	"encoding/json"
	"os"
)

type ConfigStruct struct {
	Host                       string `json:"host"`
	Port                       int    `json:"port"`
	DB                         string `json:"db"`
	CookieSecret               string `json:"cookie_secret"`
	SendMailPath               string `json:"sendmail_path"`
	SmtpUsername               string `json:"smtp_username"`
	SmtpPassword               string `json:"smtp_password"`
	SmtpHost                   string `json:"smtp_host"`
	SmtpAddr                   string `json:"smtp_addr"`
	FromEmail                  string `json:"from_email"`
	Superusers                 string `json:"superusers"`
	TimeZoneOffset             int64  `json:"time_zone_offset"`
	AnalyticsFile              string `json:"analytics_file"`
	StaticFileVersion          int    `json:"static_file_version"`
	GoGetPath                  string `json:"go_get_path"`
	PackagesDownloadPath       string `json:"packages_download_path"`
	PublicSalt                 string `json:"public_salt"`
	CookieSecure               bool   `json:"cookie_secure"`
	GithubClientId             string `json:"github_auth_client_id"`
	GithubClientSecret         string `json:"github_auth_client_secret"`
	GithubLoginRedirect        string `json:"github_login_redirect"`
	GithubLoginSuccessRedirect string `json:"github_login_success_redirect"`
	DeferPanicApiKey           string `json:"deferpanic_api_key"`
	GtCaptchaId                string `json:"gt_captcha_id"`
	GtPrivateKey               string `json:"gt_private_key"`
	GoDownloadPath             string `json:"go_download_path"`
	LiteIDEDownloadPath        string `json:"liteide_download_path"`
	ImagePath                  string `json:"image_path"`
}

var (
	Config ConfigStruct
)

func InitConfig(configFile string) error {
	return parseJsonFile("etc/config.json", &Config)
}

func parseJsonFile(path string, v interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	dec := json.NewDecoder(file)
	err = dec.Decode(v)
	if err != nil {
		return err
	}

	return nil
}
