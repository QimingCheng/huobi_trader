package untils

import (
	"testing"
	"github.com/xfrzrcj/huobi_trader/conf"
)

func Test_signed_by_jwt(t *testing.T){
	digest := "ITKa5gv8d/sfi56I/wZdi4VqLIU4GyVWO3XPxVcW+NU="
	pem := conf.PRIVATE_KEY_PRIME_256

	signed, err := SignByJWT(pem, digest)
	if nil != err {
		t.Error(err)
	} else {
		t.Log(signed, len(signed))
	}
}