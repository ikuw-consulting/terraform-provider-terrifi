package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"

	"github.com/alexklibisz/terrifi/internal/unifi"
)

func TestPortProfilePortModeToAPI(t *testing.T) {
	for _, tc := range []struct {
		mode types.String
		want *string
	}{
		{types.StringValue("edge"), ptrTo("enabled")},
		{types.StringValue("infrastructure"), ptrTo("disabled")},
		{types.StringNull(), nil},
	} {
		t.Run(tc.mode.String(), func(t *testing.T) {
			m := defaultPortProfileModel("mode")
			m.PortMode = tc.mode
			wire := toWire(t, mustModelToAPI(t, &m, nil))
			if tc.want == nil {
				assert.NotContains(t, wire, "stp_edge_state", "null port_mode must leave the controller value untouched")
				return
			}
			assert.Equal(t, *tc.want, wire["stp_edge_state"])
		})
	}
}

func TestPortProfilePortModeFromAPI(t *testing.T) {
	for _, tc := range []struct {
		api  *string
		want types.String
	}{
		{ptrTo("enabled"), types.StringValue("edge")},
		{ptrTo("disabled"), types.StringValue("infrastructure")},
		{ptrTo(""), types.StringNull()},
		{ptrTo("unexpected"), types.StringNull()},
		{nil, types.StringNull()},
	} {
		var m portProfileResourceModel
		(&portProfileResource{}).apiToModel(&unifi.PortProfile{StpEdgeState: tc.api}, &m, "default", nil)
		assert.Equal(t, tc.want, m.PortMode)
	}
}

func TestAccPortProfile_portMode(t *testing.T) {
	name := fmt.Sprintf("tfacc-pp-mode-%s", randomSuffix())
	step := func(mode string, check resource.TestCheckFunc) resource.TestStep {
		body := ""
		if mode != "" {
			body = fmt.Sprintf("  port_mode = %q", mode)
		}
		return resource.TestStep{Config: testAccPortProfileConfig(name, body), Check: check}
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheckPortProfile(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPortProfileDestroyed,
		Steps: []resource.TestStep{
			step("edge", resource.TestCheckResourceAttr(testAccPortProfileAddr, "port_mode", "edge")),
			{ResourceName: testAccPortProfileAddr, ImportState: true, ImportStateVerify: true},
			step("infrastructure", resource.TestCheckResourceAttr(testAccPortProfileAddr, "port_mode", "infrastructure")),
			{ResourceName: testAccPortProfileAddr, ImportState: true, ImportStateVerify: true},
			step("edge", resource.TestCheckResourceAttr(testAccPortProfileAddr, "port_mode", "edge")),
		},
	})
}
