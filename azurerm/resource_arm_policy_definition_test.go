package azurerm

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAzureRMPolicyDefinition_basic(t *testing.T) {
	resourceName := "azurerm_policy_definition.test"

	ri := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMPolicyDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAzureRMPolicyDefinition_basic(ri),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMPolicyDefinitionExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAzureRMPolicyDefinitionAtMgmtGroup_basic(t *testing.T) {
	resourceName := "azurerm_policy_definition.test"
	mgmtGroupName := "azurerm_management_group.test"

	ri := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMPolicyDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAzureRMPolicyDefinition_ManagementGroup(ri),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMPolicyDefinitionExistsInMgmtGroup(resourceName, mgmtGroupName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testCheckAzureRMPolicyDefinitionExistsInMgmtGroup(policyName string, managementGroupName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[policyName]
		if !ok {
			return fmt.Errorf("not found: %s", policyName)
		}

		policyName := rs.Primary.Attributes["name"]
		managementGroupID := rs.Primary.Attributes["management_group_id"]

		client := testAccProvider.Meta().(*ArmClient).policyDefinitionsClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext

		resp, err := client.GetAtManagementGroup(ctx, policyName, managementGroupID)
		if err != nil {
			return fmt.Errorf("Bad: GetAtManagementGroup on policyDefinitionsClient: %s", err)
		}

		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("policy does not exist: %s", policyName)
		}

		return nil
	}
}

func testCheckAzureRMPolicyDefinitionExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		policyName := rs.Primary.Attributes["name"]

		client := testAccProvider.Meta().(*ArmClient).policyDefinitionsClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext

		resp, err := client.Get(ctx, policyName)
		if err != nil {
			return fmt.Errorf("Bad: Get on policyDefinitionsClient: %s", err)
		}

		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("policy does not exist: %s", name)
		}

		return nil
	}
}

func testCheckAzureRMPolicyDefinitionDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*ArmClient).policyDefinitionsClient
	ctx := testAccProvider.Meta().(*ArmClient).StopContext

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurerm_policy_definition" {
			continue
		}

		name := rs.Primary.Attributes["name"]

		resp, err := client.Get(ctx, name)

		if err != nil {
			return nil
		}

		if resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("policy still exists:%s", *resp.Name)
		}
	}

	return nil
}

func testAzureRMPolicyDefinition_basic(ri int) string {
	return fmt.Sprintf(`
resource "azurerm_policy_definition" "test" {
  name         = "acctestpol-%d"
  policy_type  = "Custom"
  mode         = "All"
  display_name = "acctestpol-%d"

  policy_rule = <<POLICY_RULE
	{
    "if": {
      "not": {
        "field": "location",
        "in": "[parameters('allowedLocations')]"
      }
    },
    "then": {
      "effect": "audit"
    }
  }
POLICY_RULE

  parameters = <<PARAMETERS
	{
    "allowedLocations": {
      "type": "Array",
      "metadata": {
        "description": "The list of allowed locations for resources.",
        "displayName": "Allowed locations",
        "strongType": "location"
      }
    }
  }
PARAMETERS
}
`, ri, ri)
}

func testAzureRMPolicyDefinition_ManagementGroup(ri int) string {
	return fmt.Sprintf(`
resource "azurerm_management_group" "test" {
	display_name = "acctestmg-%d"
}

resource "azurerm_policy_definition" "test" {
  name         = "acctestpol-%d"
  policy_type  = "Custom"
  mode         = "All"
  display_name = "acctestpol-%d"
  management_group_id = "${azurerm_management_group.test.group_id}"

  policy_rule = <<POLICY_RULE
	{
    "if": {
      "not": {
        "field": "location",
        "in": "[parameters('allowedLocations')]"
      }
    },
    "then": {
      "effect": "audit"
    }
  }
POLICY_RULE

  parameters = <<PARAMETERS
	{
    "allowedLocations": {
      "type": "Array",
      "metadata": {
        "description": "The list of allowed locations for resources.",
        "displayName": "Allowed locations",
        "strongType": "location"
      }
    }
  }
PARAMETERS
}
`, ri, ri, ri)
}
