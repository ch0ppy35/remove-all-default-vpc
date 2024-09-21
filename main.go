package main

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func getRegions(ctx context.Context, client EC2API) ([]string, error) {
	output, err := client.DescribeRegions(ctx, &ec2.DescribeRegionsInput{})
	if err != nil {
		return nil, err
	}

	if len(output.Regions) == 0 {
		return []string{}, nil
	}

	regions := []string{}
	for _, region := range output.Regions {
		regions = append(regions, *region.RegionName)
	}

	return regions, nil
}

func getDefaultVPCs(ctx context.Context, client EC2API) ([]string, error) {
	output, err := client.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{})
	if err != nil {
		return nil, err
	}

	if len(output.Vpcs) == 0 {
		return []string{}, nil
	}

	vpcs := []string{}
	for _, vpc := range output.Vpcs {
		if *vpc.IsDefault {
			vpcs = append(vpcs, *vpc.VpcId)
		}
	}

	return vpcs, nil
}

// Delete subnets in a VPC
func deleteSubnets(ctx context.Context, client EC2API, vpcID string) error {
	resp, err := client.DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []string{vpcID},
			},
		},
	})
	if err != nil {
		return err
	}

	for _, subnet := range resp.Subnets {
		_, err := client.DeleteSubnet(ctx, &ec2.DeleteSubnetInput{
			SubnetId: subnet.SubnetId,
		})
		if err != nil {
			return fmt.Errorf("failed to delete subnet %s: %w", aws.ToString(subnet.SubnetId), err)
		}
		fmt.Printf("Deleted subnet: %s\n", aws.ToString(subnet.SubnetId))
	}
	return nil
}

func isMainRouteTable(rt types.RouteTable) bool {
	for _, association := range rt.Associations {
		if association.Main != nil && *association.Main {
			return true
		}
	}
	return false
}

// Delete route tables in a VPC
func deleteRouteTables(ctx context.Context, client EC2API, vpcID string) error {
	resp, err := client.DescribeRouteTables(ctx, &ec2.DescribeRouteTablesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []string{vpcID},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to describe route tables: %w", err)
	}

	for _, rt := range resp.RouteTables {
		if isMainRouteTable(rt) {
			continue
		}

		_, err := client.DeleteRouteTable(ctx, &ec2.DeleteRouteTableInput{
			RouteTableId: rt.RouteTableId,
		})
		if err != nil {
			return fmt.Errorf("failed to delete route table %s: %w", aws.ToString(rt.RouteTableId), err)
		}
		fmt.Printf("Deleted route table: %s\n", aws.ToString(rt.RouteTableId))
	}
	return nil
}

// Detach and delete internet gateways in a VPC
func deleteInternetGateways(ctx context.Context, client EC2API, vpcID string) error {
	resp, err := client.DescribeInternetGateways(ctx, &ec2.DescribeInternetGatewaysInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("attachment.vpc-id"),
				Values: []string{vpcID},
			},
		},
	})
	if err != nil {
		return err
	}

	for _, igw := range resp.InternetGateways {
		_, err := client.DetachInternetGateway(ctx, &ec2.DetachInternetGatewayInput{
			InternetGatewayId: igw.InternetGatewayId,
			VpcId:             aws.String(vpcID),
		})
		if err != nil {
			return fmt.Errorf("failed to detach internet gateway %s: %w", aws.ToString(igw.InternetGatewayId), err)
		}

		_, err = client.DeleteInternetGateway(ctx, &ec2.DeleteInternetGatewayInput{
			InternetGatewayId: igw.InternetGatewayId,
		})
		if err != nil {
			return fmt.Errorf("failed to delete internet gateway %s: %w", aws.ToString(igw.InternetGatewayId), err)
		}
		fmt.Printf("Deleted internet gateway: %s\n", aws.ToString(igw.InternetGatewayId))
	}
	return nil
}

// Delete security groups in a VPC
func deleteSecurityGroups(ctx context.Context, client EC2API, vpcID string) error {
	resp, err := client.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []string{vpcID},
			},
		},
	})

	if err != nil {
		return err
	}

	for _, sg := range resp.SecurityGroups {
		if *sg.GroupName == "default" {
			continue
		}
		_, err := client.DeleteSecurityGroup(ctx, &ec2.DeleteSecurityGroupInput{
			GroupId: sg.GroupId,
		})
		if err != nil {
			return fmt.Errorf("failed to delete security group %s: %w", aws.ToString(sg.GroupId), err)
		}
		fmt.Printf("Deleted security group: %s\n", aws.ToString(sg.GroupId))
	}
	return nil
}

// Delete network ACLs in a VPC
func deleteNetworkACLs(ctx context.Context, client EC2API, vpcID string) error {
	resp, err := client.DescribeNetworkAcls(ctx, &ec2.DescribeNetworkAclsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []string{vpcID},
			},
		},
	})
	if err != nil {
		return err
	}

	for _, acl := range resp.NetworkAcls {
		if *acl.IsDefault {
			continue
		}
		_, err := client.DeleteNetworkAcl(ctx, &ec2.DeleteNetworkAclInput{
			NetworkAclId: acl.NetworkAclId,
		})
		if err != nil {
			return fmt.Errorf("failed to delete network ACL %s: %w", aws.ToString(acl.NetworkAclId), err)
		}
		fmt.Printf("Deleted network ACL: %s\n", aws.ToString(acl.NetworkAclId))
	}
	return nil
}

// Delete the VPC after cleaning up resources
func deleteVPC(ctx context.Context, client EC2API, vpcID string) error {
	_, err := client.DeleteVpc(ctx, &ec2.DeleteVpcInput{
		VpcId: aws.String(vpcID),
	})
	if err != nil {
		return fmt.Errorf("failed to delete VPC %s: %w", vpcID, err)
	}

	fmt.Printf("Deleted VPC: %s\n", vpcID)
	return nil
}

// Clean up resources in a VPC before deleting it
func cleanupVPCResources(ctx context.Context, client EC2API, vpcID string) error {
	err := deleteInternetGateways(ctx, client, vpcID)
	if err != nil {
		return err
	}

	err = deleteSubnets(ctx, client, vpcID)
	if err != nil {
		return err
	}

	err = deleteRouteTables(ctx, client, vpcID)
	if err != nil {
		return err
	}

	err = deleteNetworkACLs(ctx, client, vpcID)
	if err != nil {
		return err
	}

	err = deleteSecurityGroups(ctx, client, vpcID)
	if err != nil {
		return err
	}
	return nil
}

// deleteAllDefaultVPCs deletes all default VPCs in all regions
func DeleteAllDefaultVPCs(ctx context.Context, regions []string, cfg aws.Config) {
	var wg sync.WaitGroup
	for _, region := range regions {
		wg.Add(1)
		go func(region string) {
			defer wg.Done()
			fmt.Printf("Processing region: %s\n", region)
			regionCfg := cfg.Copy()
			regionCfg.Region = region

			ec2Client := &EC2Client{Client: ec2.NewFromConfig(regionCfg)}

			vpcs, err := getDefaultVPCs(ctx, ec2Client)
			if err != nil {
				fmt.Printf("Error fetching default VPCs in region %s: %v\n", region, err)
				return
			}

			for _, vpcID := range vpcs {
				err := cleanupVPCResources(ctx, ec2Client, vpcID)
				if err != nil {
					fmt.Printf("Error cleaning up resources for VPC %s: %v", vpcID, err)
					os.Exit(1)
				}

				fmt.Printf("Deleting default VPC %s in region %s\n", vpcID, region)
				err = deleteVPC(ctx, ec2Client, vpcID)
				if err != nil {
					fmt.Printf("Error deleting VPC %s in region %s: %v", vpcID, region, err)
					os.Exit(1)
				}
			}
		}(region)
	}

	wg.Wait()
	fmt.Println("All default VPCs deleted.")
}

func main() {
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		fmt.Printf("Unable to load AWS SDK config: %v", err)
		os.Exit(1)
	}

	ec2Client := &EC2Client{Client: ec2.NewFromConfig(cfg)}

	regions, err := getRegions(ctx, ec2Client)
	if err != nil {
		fmt.Printf("Unable to describe regions: %v", err)
		os.Exit(1)
	}

	DeleteAllDefaultVPCs(ctx, regions, cfg)
}
