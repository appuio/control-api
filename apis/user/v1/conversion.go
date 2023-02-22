package v1

import (
	"net/url"

	"k8s.io/apimachinery/pkg/conversion"
)

func Convert_url_Values_To__RedeemOptions(in, out interface{}, _ conversion.Scope) error {
	return convert_url_Values_To__RedeemOptions(in.(*url.Values), out.(*RedeemOptions))
}

func convert_url_Values_To__RedeemOptions(in *url.Values, out *RedeemOptions) error {
	out.Token = in.Get("token")
	return nil
}
