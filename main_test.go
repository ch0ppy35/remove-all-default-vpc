package main

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func Test_getRegions(t *testing.T) {
	ctx := context.TODO()

	tests := []struct {
		name    string
		client  EC2API
		want    []string
		wantErr bool
	}{
		{
			name: "success - multiple regions",
			client: &MockEC2Client{
				describeRegionsFunc: func(ctx context.Context, input *ec2.DescribeRegionsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeRegionsOutput, error) {
					return &ec2.DescribeRegionsOutput{
						Regions: []types.Region{
							{RegionName: aws.String("us-east-1")},
							{RegionName: aws.String("us-west-2")},
						},
					}, nil
				},
			},
			want:    []string{"us-east-1", "us-west-2"},
			wantErr: false,
		},
		{
			name: "error fetching regions",
			client: &MockEC2Client{
				describeRegionsFunc: func(ctx context.Context, input *ec2.DescribeRegionsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeRegionsOutput, error) {
					return nil, fmt.Errorf("failed to fetch regions")
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "no regions found",
			client: &MockEC2Client{
				describeRegionsFunc: func(ctx context.Context, input *ec2.DescribeRegionsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeRegionsOutput, error) {
					return &ec2.DescribeRegionsOutput{
						Regions: []types.Region{},
					}, nil
				},
			},
			want:    []string{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getRegions(ctx, tt.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("getRegions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getRegions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getDefaultVPCs(t *testing.T) {
	ctx := context.TODO()

	tests := []struct {
		name    string
		client  EC2API
		want    []string
		wantErr bool
	}{
		{
			name: "success - multiple default VPCs",
			client: &MockEC2Client{
				describeVpcsFunc: func(ctx context.Context, input *ec2.DescribeVpcsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVpcsOutput, error) {
					return &ec2.DescribeVpcsOutput{
						Vpcs: []types.Vpc{
							{VpcId: aws.String("vpc-12345"), IsDefault: aws.Bool(true)},
							{VpcId: aws.String("vpc-67890"), IsDefault: aws.Bool(true)},
						},
					}, nil
				},
			},
			want:    []string{"vpc-12345", "vpc-67890"},
			wantErr: false,
		},
		{
			name: "error fetching VPCs",
			client: &MockEC2Client{
				describeVpcsFunc: func(ctx context.Context, input *ec2.DescribeVpcsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVpcsOutput, error) {
					return nil, fmt.Errorf("failed to fetch VPCs")
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "no default VPCs",
			client: &MockEC2Client{
				describeVpcsFunc: func(ctx context.Context, input *ec2.DescribeVpcsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVpcsOutput, error) {
					return &ec2.DescribeVpcsOutput{
						Vpcs: []types.Vpc{},
					}, nil
				},
			},
			want:    []string{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getDefaultVPCs(ctx, tt.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("getDefaultVPCs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getDefaultVPCs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_deleteSubnets(t *testing.T) {
	ctx := context.TODO()

	tests := []struct {
		name    string
		client  EC2API
		vpcID   string
		wantErr bool
	}{
		{
			name: "success - delete multiple subnets",
			client: &MockEC2Client{
				describeSubnetsFunc: func(ctx context.Context, input *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error) {
					return &ec2.DescribeSubnetsOutput{
						Subnets: []types.Subnet{
							{SubnetId: aws.String("subnet-1")},
							{SubnetId: aws.String("subnet-2")},
						},
					}, nil
				},
				deleteSubnetFunc: func(ctx context.Context, input *ec2.DeleteSubnetInput, optFns ...func(*ec2.Options)) (*ec2.DeleteSubnetOutput, error) {
					return &ec2.DeleteSubnetOutput{}, nil
				},
			},
			vpcID:   "vpc-12345",
			wantErr: false,
		},
		{
			name: "error fetching subnets",
			client: &MockEC2Client{
				describeSubnetsFunc: func(ctx context.Context, input *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error) {
					return nil, fmt.Errorf("failed to fetch subnets")
				},
			},
			vpcID:   "vpc-12345",
			wantErr: true,
		},
		{
			name: "error deleting subnet",
			client: &MockEC2Client{
				describeSubnetsFunc: func(ctx context.Context, input *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error) {
					return &ec2.DescribeSubnetsOutput{
						Subnets: []types.Subnet{
							{SubnetId: aws.String("subnet-1")},
						},
					}, nil
				},
				deleteSubnetFunc: func(ctx context.Context, input *ec2.DeleteSubnetInput, optFns ...func(*ec2.Options)) (*ec2.DeleteSubnetOutput, error) {
					return nil, fmt.Errorf("failed to delete subnet subnet-1")
				},
			},
			vpcID:   "vpc-12345",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := deleteSubnets(ctx, tt.client, tt.vpcID)
			if (err != nil) != tt.wantErr {
				t.Errorf("deleteSubnets() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_deleteRouteTables(t *testing.T) {
	type args struct {
		ctx    context.Context
		client EC2API
		vpcID  string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success - delete multiple route tables",
			args: args{
				ctx: context.TODO(),
				client: &MockEC2Client{
					describeRouteTablesFunc: func(ctx context.Context, input *ec2.DescribeRouteTablesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeRouteTablesOutput, error) {
						return &ec2.DescribeRouteTablesOutput{
							RouteTables: []types.RouteTable{
								{RouteTableId: aws.String("rtb-1")},
								{RouteTableId: aws.String("rtb-2")},
							},
						}, nil
					},
					deleteRouteTableFunc: func(ctx context.Context, input *ec2.DeleteRouteTableInput, optFns ...func(*ec2.Options)) (*ec2.DeleteRouteTableOutput, error) {
						return &ec2.DeleteRouteTableOutput{}, nil
					},
				},
				vpcID: "vpc-12345",
			},
			wantErr: false,
		},
		{
			name: "error describing route tables",
			args: args{
				ctx: context.TODO(),
				client: &MockEC2Client{
					describeRouteTablesFunc: func(ctx context.Context, input *ec2.DescribeRouteTablesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeRouteTablesOutput, error) {
						return nil, fmt.Errorf("failed to describe route tables")
					},
				},
				vpcID: "vpc-12345",
			},
			wantErr: true,
		},
		{
			name: "error deleting route table",
			args: args{
				ctx: context.TODO(),
				client: &MockEC2Client{
					describeRouteTablesFunc: func(ctx context.Context, input *ec2.DescribeRouteTablesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeRouteTablesOutput, error) {
						return &ec2.DescribeRouteTablesOutput{
							RouteTables: []types.RouteTable{
								{RouteTableId: aws.String("rtb-1")},
							},
						}, nil
					},
					deleteRouteTableFunc: func(ctx context.Context, input *ec2.DeleteRouteTableInput, optFns ...func(*ec2.Options)) (*ec2.DeleteRouteTableOutput, error) {
						return nil, fmt.Errorf("failed to delete route table")
					},
				},
				vpcID: "vpc-12345",
			},
			wantErr: true,
		},
		{
			name: "no route tables found",
			args: args{
				ctx: context.TODO(),
				client: &MockEC2Client{
					describeRouteTablesFunc: func(ctx context.Context, input *ec2.DescribeRouteTablesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeRouteTablesOutput, error) {
						return &ec2.DescribeRouteTablesOutput{
							RouteTables: []types.RouteTable{},
						}, nil
					},
				},
				vpcID: "vpc-12345",
			},
			wantErr: false,
		},
		{
			name: "success - skip main route table",
			args: args{
				ctx: context.TODO(),
				client: &MockEC2Client{
					describeRouteTablesFunc: func(ctx context.Context, input *ec2.DescribeRouteTablesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeRouteTablesOutput, error) {
						return &ec2.DescribeRouteTablesOutput{
							RouteTables: []types.RouteTable{
								{
									RouteTableId: aws.String("rtb-1"),
									Associations: []types.RouteTableAssociation{
										{
											Main: aws.Bool(true),
										},
									},
								},
								{
									RouteTableId: aws.String("rtb-2"),
									Associations: []types.RouteTableAssociation{
										{
											Main: aws.Bool(false),
										},
									},
								},
							},
						}, nil
					},
					deleteRouteTableFunc: func(ctx context.Context, input *ec2.DeleteRouteTableInput, optFns ...func(*ec2.Options)) (*ec2.DeleteRouteTableOutput, error) {
						if *input.RouteTableId == "rtb-2" {
							return &ec2.DeleteRouteTableOutput{}, nil
						}
						return nil, fmt.Errorf("failed to delete route table")
					},
				},
				vpcID: "vpc-12345",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := deleteRouteTables(tt.args.ctx, tt.args.client, tt.args.vpcID); (err != nil) != tt.wantErr {
				t.Errorf("deleteRouteTables() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_deleteInternetGateways(t *testing.T) {
	type args struct {
		ctx    context.Context
		client EC2API
		vpcID  string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success - multiple internet gateways",
			args: args{
				ctx: context.TODO(),
				client: &MockEC2Client{
					describeInternetGatewaysFunc: func(ctx context.Context, input *ec2.DescribeInternetGatewaysInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInternetGatewaysOutput, error) {
						return &ec2.DescribeInternetGatewaysOutput{
							InternetGateways: []types.InternetGateway{
								{InternetGatewayId: aws.String("igw-12345")},
								{InternetGatewayId: aws.String("igw-67890")},
							},
						}, nil
					},
					detachInternetGatewayFunc: func(ctx context.Context, input *ec2.DetachInternetGatewayInput, optFns ...func(*ec2.Options)) (*ec2.DetachInternetGatewayOutput, error) {
						return &ec2.DetachInternetGatewayOutput{}, nil
					},
					deleteInternetGatewayFunc: func(ctx context.Context, input *ec2.DeleteInternetGatewayInput, optFns ...func(*ec2.Options)) (*ec2.DeleteInternetGatewayOutput, error) {
						return &ec2.DeleteInternetGatewayOutput{}, nil
					},
				},
				vpcID: "vpc-12345",
			},
			wantErr: false,
		},
		{
			name: "error detaching internet gateway",
			args: args{
				ctx: context.TODO(),
				client: &MockEC2Client{
					describeInternetGatewaysFunc: func(ctx context.Context, input *ec2.DescribeInternetGatewaysInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInternetGatewaysOutput, error) {
						return &ec2.DescribeInternetGatewaysOutput{
							InternetGateways: []types.InternetGateway{
								{InternetGatewayId: aws.String("igw-12345")},
							},
						}, nil
					},
					detachInternetGatewayFunc: func(ctx context.Context, input *ec2.DetachInternetGatewayInput, optFns ...func(*ec2.Options)) (*ec2.DetachInternetGatewayOutput, error) {
						return nil, fmt.Errorf("failed to detach internet gateway")
					},
				},
				vpcID: "vpc-12345",
			},
			wantErr: true,
		},
		{
			name: "error deleting internet gateway",
			args: args{
				ctx: context.TODO(),
				client: &MockEC2Client{
					describeInternetGatewaysFunc: func(ctx context.Context, input *ec2.DescribeInternetGatewaysInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInternetGatewaysOutput, error) {
						return &ec2.DescribeInternetGatewaysOutput{
							InternetGateways: []types.InternetGateway{
								{InternetGatewayId: aws.String("igw-12345")},
							},
						}, nil
					},
					detachInternetGatewayFunc: func(ctx context.Context, input *ec2.DetachInternetGatewayInput, optFns ...func(*ec2.Options)) (*ec2.DetachInternetGatewayOutput, error) {
						return &ec2.DetachInternetGatewayOutput{}, nil
					},
					deleteInternetGatewayFunc: func(ctx context.Context, input *ec2.DeleteInternetGatewayInput, optFns ...func(*ec2.Options)) (*ec2.DeleteInternetGatewayOutput, error) {
						return nil, fmt.Errorf("failed to delete internet gateway")
					},
				},
				vpcID: "vpc-12345",
			},
			wantErr: true,
		},
		{
			name: "no internet gateways found",
			args: args{
				ctx: context.TODO(),
				client: &MockEC2Client{
					describeInternetGatewaysFunc: func(ctx context.Context, input *ec2.DescribeInternetGatewaysInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInternetGatewaysOutput, error) {
						return &ec2.DescribeInternetGatewaysOutput{
							InternetGateways: []types.InternetGateway{},
						}, nil
					},
				},
				vpcID: "vpc-12345",
			},
			wantErr: false,
		},
		{
			name: "error describing internet gateways",
			args: args{
				ctx: context.TODO(),
				client: &MockEC2Client{
					describeInternetGatewaysFunc: func(ctx context.Context, input *ec2.DescribeInternetGatewaysInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInternetGatewaysOutput, error) {
						return nil, fmt.Errorf("failed to describe internet gateways")
					},
				},
				vpcID: "vpc-12345",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := deleteInternetGateways(tt.args.ctx, tt.args.client, tt.args.vpcID); (err != nil) != tt.wantErr {
				t.Errorf("deleteInternetGateways() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_deleteSecurityGroups(t *testing.T) {
	type args struct {
		ctx    context.Context
		client EC2API
		vpcID  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Successfully delete a security group",
			args: args{
				ctx: context.TODO(),
				client: &MockEC2Client{
					describeSecurityGroupsFunc: func(ctx context.Context, input *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error) {
						return &ec2.DescribeSecurityGroupsOutput{
							SecurityGroups: []types.SecurityGroup{
								{
									GroupId:   aws.String("sg-12345"),
									GroupName: aws.String("test-group"),
								},
							},
						}, nil
					},
					deleteSecurityGroupFunc: func(ctx context.Context, input *ec2.DeleteSecurityGroupInput, optFns ...func(*ec2.Options)) (*ec2.DeleteSecurityGroupOutput, error) {
						return &ec2.DeleteSecurityGroupOutput{}, nil
					},
				},
				vpcID: "vpc-12345",
			},
			wantErr: false,
		},
		{
			name: "Successfully delete multiple security groups",
			args: args{
				ctx: context.TODO(),
				client: &MockEC2Client{
					describeSecurityGroupsFunc: func(ctx context.Context, input *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error) {
						return &ec2.DescribeSecurityGroupsOutput{
							SecurityGroups: []types.SecurityGroup{
								{
									GroupId:   aws.String("sg-12345"),
									GroupName: aws.String("test-group-1"),
								},
								{
									GroupId:   aws.String("sg-67890"),
									GroupName: aws.String("test-group-2"),
								},
							},
						}, nil
					},
					deleteSecurityGroupFunc: func(ctx context.Context, input *ec2.DeleteSecurityGroupInput, optFns ...func(*ec2.Options)) (*ec2.DeleteSecurityGroupOutput, error) {
						if *input.GroupId == "sg-12345" || *input.GroupId == "sg-67890" {
							return &ec2.DeleteSecurityGroupOutput{}, nil
						}
						t.Fatalf("Unexpected security group deletion attempt for group: %s", aws.ToString(input.GroupId))
						return nil, nil
					},
				},
				vpcID: "vpc-12345",
			},
			wantErr: false,
		},
		{
			name: "Skip Default Security Group",
			args: args{
				ctx: context.TODO(),
				client: &MockEC2Client{
					describeSecurityGroupsFunc: func(ctx context.Context, input *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error) {
						return &ec2.DescribeSecurityGroupsOutput{
							SecurityGroups: []types.SecurityGroup{
								{
									GroupId:   aws.String("sg-default"),
									GroupName: aws.String("default"),
								},
							},
						}, nil
					},
					deleteSecurityGroupFunc: func(ctx context.Context, input *ec2.DeleteSecurityGroupInput, optFns ...func(*ec2.Options)) (*ec2.DeleteSecurityGroupOutput, error) {
						t.Fatalf("deleteSecurityGroupFunc should not be called for the default group")
						return nil, nil
					},
				},
				vpcID: "vpc-12345",
			},
			wantErr: false,
		},
		{
			name: "No Security Groups to Delete",
			args: args{
				ctx: context.TODO(),
				client: &MockEC2Client{
					describeSecurityGroupsFunc: func(ctx context.Context, input *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error) {
						return &ec2.DescribeSecurityGroupsOutput{
							SecurityGroups: []types.SecurityGroup{},
						}, nil
					},
				},
				vpcID: "vpc-12345",
			},
			wantErr: false,
		},
		{
			name: "Error in DescribeSecurityGroups",
			args: args{
				ctx: context.TODO(),
				client: &MockEC2Client{
					describeSecurityGroupsFunc: func(ctx context.Context, input *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error) {
						return nil, errors.New("describe security groups failed")
					},
				},
				vpcID: "vpc-12345",
			},
			wantErr: true,
		},
		{
			name: "Error in DeleteSecurityGroup",
			args: args{
				ctx: context.TODO(),
				client: &MockEC2Client{
					describeSecurityGroupsFunc: func(ctx context.Context, input *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error) {
						return &ec2.DescribeSecurityGroupsOutput{
							SecurityGroups: []types.SecurityGroup{
								{
									GroupId:   aws.String("sg-12345"),
									GroupName: aws.String("test-group"),
								},
							},
						}, nil
					},
					deleteSecurityGroupFunc: func(ctx context.Context, input *ec2.DeleteSecurityGroupInput, optFns ...func(*ec2.Options)) (*ec2.DeleteSecurityGroupOutput, error) {
						return nil, errors.New("delete security group failed")
					},
				},
				vpcID: "vpc-12345",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := deleteSecurityGroups(tt.args.ctx, tt.args.client, tt.args.vpcID); (err != nil) != tt.wantErr {
				t.Errorf("deleteSecurityGroups() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_deleteNetworkACLs(t *testing.T) {
	type args struct {
		ctx    context.Context
		client EC2API
		vpcID  string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Successfully delete non-default network ACLs",
			args: args{
				ctx: context.TODO(),
				client: &MockEC2Client{
					describeNetworkAclsFunc: func(ctx context.Context, input *ec2.DescribeNetworkAclsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeNetworkAclsOutput, error) {
						return &ec2.DescribeNetworkAclsOutput{
							NetworkAcls: []types.NetworkAcl{
								{
									NetworkAclId: aws.String("acl-12345"),
									IsDefault:    aws.Bool(false),
								},
							},
						}, nil
					},
					deleteNetworkAclFunc: func(ctx context.Context, input *ec2.DeleteNetworkAclInput, optFns ...func(*ec2.Options)) (*ec2.DeleteNetworkAclOutput, error) {
						return &ec2.DeleteNetworkAclOutput{}, nil
					},
				},
				vpcID: "vpc-12345",
			},
			wantErr: false,
		},
		{
			name: "Successfully skip default network ACL",
			args: args{
				ctx: context.TODO(),
				client: &MockEC2Client{
					describeNetworkAclsFunc: func(ctx context.Context, input *ec2.DescribeNetworkAclsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeNetworkAclsOutput, error) {
						return &ec2.DescribeNetworkAclsOutput{
							NetworkAcls: []types.NetworkAcl{
								{
									NetworkAclId: aws.String("acl-default"),
									IsDefault:    aws.Bool(true),
								},
							},
						}, nil
					},
				},
				vpcID: "vpc-12345",
			},
			wantErr: false,
		},
		{
			name: "Error describing network ACLs",
			args: args{
				ctx: context.TODO(),
				client: &MockEC2Client{
					describeNetworkAclsFunc: func(ctx context.Context, input *ec2.DescribeNetworkAclsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeNetworkAclsOutput, error) {
						return nil, fmt.Errorf("failed to describe network ACLs")
					},
				},
				vpcID: "vpc-12345",
			},
			wantErr: true,
		},
		{
			name: "Error deleting a network ACL",
			args: args{
				ctx: context.TODO(),
				client: &MockEC2Client{
					describeNetworkAclsFunc: func(ctx context.Context, input *ec2.DescribeNetworkAclsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeNetworkAclsOutput, error) {
						return &ec2.DescribeNetworkAclsOutput{
							NetworkAcls: []types.NetworkAcl{
								{
									NetworkAclId: aws.String("acl-12345"),
									IsDefault:    aws.Bool(false),
								},
							},
						}, nil
					},
					deleteNetworkAclFunc: func(ctx context.Context, input *ec2.DeleteNetworkAclInput, optFns ...func(*ec2.Options)) (*ec2.DeleteNetworkAclOutput, error) {
						return nil, fmt.Errorf("failed to delete network ACL")
					},
				},
				vpcID: "vpc-12345",
			},
			wantErr: true,
		},
		{
			name: "No network ACLs found",
			args: args{
				ctx: context.TODO(),
				client: &MockEC2Client{
					describeNetworkAclsFunc: func(ctx context.Context, input *ec2.DescribeNetworkAclsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeNetworkAclsOutput, error) {
						return &ec2.DescribeNetworkAclsOutput{
							NetworkAcls: []types.NetworkAcl{},
						}, nil
					},
				},
				vpcID: "vpc-12345",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := deleteNetworkACLs(tt.args.ctx, tt.args.client, tt.args.vpcID); (err != nil) != tt.wantErr {
				t.Errorf("deleteNetworkACLs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_deleteVPC(t *testing.T) {
	type args struct {
		ctx    context.Context
		client EC2API
		vpcID  string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Successfully delete VPC",
			args: args{
				ctx: context.TODO(),
				client: &MockEC2Client{
					deleteVpcFunc: func(ctx context.Context, input *ec2.DeleteVpcInput, optFns ...func(*ec2.Options)) (*ec2.DeleteVpcOutput, error) {
						return &ec2.DeleteVpcOutput{}, nil
					},
				},
				vpcID: "vpc-12345",
			},
			wantErr: false,
		},
		{
			name: "Error deleting VPC",
			args: args{
				ctx: context.TODO(),
				client: &MockEC2Client{
					deleteVpcFunc: func(ctx context.Context, input *ec2.DeleteVpcInput, optFns ...func(*ec2.Options)) (*ec2.DeleteVpcOutput, error) {
						return nil, fmt.Errorf("failed to delete VPC")
					},
				},
				vpcID: "vpc-12345",
			},
			wantErr: true,
		},
		{
			name: "Delete non-existent VPC",
			args: args{
				ctx: context.TODO(),
				client: &MockEC2Client{
					deleteVpcFunc: func(ctx context.Context, input *ec2.DeleteVpcInput, optFns ...func(*ec2.Options)) (*ec2.DeleteVpcOutput, error) {
						return nil, fmt.Errorf("VPC not found")
					},
				},
				vpcID: "vpc-non-existent",
			},
			wantErr: true,
		},
		{
			name: "Error deleting due to dependency",
			args: args{
				ctx: context.TODO(),
				client: &MockEC2Client{
					deleteVpcFunc: func(ctx context.Context, input *ec2.DeleteVpcInput, optFns ...func(*ec2.Options)) (*ec2.DeleteVpcOutput, error) {
						return nil, fmt.Errorf("DependencyViolation: VPC has dependent resources")
					},
				},
				vpcID: "vpc-12345",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := deleteVPC(tt.args.ctx, tt.args.client, tt.args.vpcID); (err != nil) != tt.wantErr {
				t.Errorf("deleteVPC() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_cleanupVPCResources(t *testing.T) {
	type args struct {
		ctx    context.Context
		client EC2API
		vpcID  string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Successful cleanup",
			args: args{
				ctx: context.Background(),
				client: &MockEC2Client{
					describeInternetGatewaysFunc: func(ctx context.Context, input *ec2.DescribeInternetGatewaysInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInternetGatewaysOutput, error) {
						return &ec2.DescribeInternetGatewaysOutput{
							InternetGateways: []types.InternetGateway{
								{
									InternetGatewayId: aws.String("igw-12345"),
								},
							},
						}, nil
					},
					describeSubnetsFunc: func(ctx context.Context, input *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error) {
						return &ec2.DescribeSubnetsOutput{
							Subnets: []types.Subnet{
								{
									SubnetId: aws.String("subnet-1"),
								},
								{
									SubnetId: aws.String("subnet-2"),
								},
							},
						}, nil
					},
					describeRouteTablesFunc: func(ctx context.Context, input *ec2.DescribeRouteTablesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeRouteTablesOutput, error) {
						return &ec2.DescribeRouteTablesOutput{
							RouteTables: []types.RouteTable{
								{RouteTableId: aws.String("rtb-1")},
								{RouteTableId: aws.String("rtb-2")},
							},
						}, nil
					},
					describeSecurityGroupsFunc: func(ctx context.Context, input *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error) {
						return &ec2.DescribeSecurityGroupsOutput{
							SecurityGroups: []types.SecurityGroup{
								{
									GroupId:   aws.String("sg-12345"),
									GroupName: aws.String("test-group-0"),
								},
								{
									GroupId:   aws.String("sg-67890"),
									GroupName: aws.String("test-group-1"),
								},
							},
						}, nil
					},
					describeNetworkAclsFunc: func(ctx context.Context, input *ec2.DescribeNetworkAclsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeNetworkAclsOutput, error) {
						return &ec2.DescribeNetworkAclsOutput{
							NetworkAcls: []types.NetworkAcl{
								{
									IsDefault:    aws.Bool(false),
									NetworkAclId: aws.String("acl-12345"),
								},
								{
									IsDefault:    aws.Bool(false),
									NetworkAclId: aws.String("acl-67890"),
								},
							},
						}, nil
					},
					deleteInternetGatewayFunc: func(ctx context.Context, input *ec2.DeleteInternetGatewayInput, optFns ...func(*ec2.Options)) (*ec2.DeleteInternetGatewayOutput, error) {
						return &ec2.DeleteInternetGatewayOutput{}, nil
					},
					deleteSubnetFunc: func(ctx context.Context, input *ec2.DeleteSubnetInput, optFns ...func(*ec2.Options)) (*ec2.DeleteSubnetOutput, error) {
						return &ec2.DeleteSubnetOutput{}, nil
					},
					deleteRouteTableFunc: func(ctx context.Context, input *ec2.DeleteRouteTableInput, optFns ...func(*ec2.Options)) (*ec2.DeleteRouteTableOutput, error) {
						return &ec2.DeleteRouteTableOutput{}, nil
					},
					deleteSecurityGroupFunc: func(ctx context.Context, input *ec2.DeleteSecurityGroupInput, optFns ...func(*ec2.Options)) (*ec2.DeleteSecurityGroupOutput, error) {
						return &ec2.DeleteSecurityGroupOutput{}, nil
					},
					deleteNetworkAclFunc: func(ctx context.Context, input *ec2.DeleteNetworkAclInput, optFns ...func(*ec2.Options)) (*ec2.DeleteNetworkAclOutput, error) {
						return &ec2.DeleteNetworkAclOutput{}, nil
					},
					detachInternetGatewayFunc: func(ctx context.Context, input *ec2.DetachInternetGatewayInput, optFns ...func(*ec2.Options)) (*ec2.DetachInternetGatewayOutput, error) {
						return &ec2.DetachInternetGatewayOutput{}, nil
					},
				},
				vpcID: "vpc-12345",
			},
			wantErr: false,
		},
		{
			name: "Error deleting internet gateways",
			args: args{
				ctx: context.Background(),
				client: &MockEC2Client{
					describeInternetGatewaysFunc: func(ctx context.Context, input *ec2.DescribeInternetGatewaysInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInternetGatewaysOutput, error) {
						return &ec2.DescribeInternetGatewaysOutput{
							InternetGateways: []types.InternetGateway{
								{
									InternetGatewayId: aws.String("igw-12345"),
								},
							},
						}, nil
					},
					detachInternetGatewayFunc: func(ctx context.Context, input *ec2.DetachInternetGatewayInput, optFns ...func(*ec2.Options)) (*ec2.DetachInternetGatewayOutput, error) {
						return &ec2.DetachInternetGatewayOutput{}, nil
					},
					deleteInternetGatewayFunc: func(ctx context.Context, input *ec2.DeleteInternetGatewayInput, optFns ...func(*ec2.Options)) (*ec2.DeleteInternetGatewayOutput, error) {
						return nil, fmt.Errorf("failed to delete internet gateway")
					},
				},
				vpcID: "vpc-12345",
			},
			wantErr: true,
		},
		{
			name: "Error deleting subnets",
			args: args{
				ctx: context.Background(),
				client: &MockEC2Client{
					deleteInternetGatewayFunc: func(ctx context.Context, input *ec2.DeleteInternetGatewayInput, optFns ...func(*ec2.Options)) (*ec2.DeleteInternetGatewayOutput, error) {
						return &ec2.DeleteInternetGatewayOutput{}, nil
					},
					deleteSubnetFunc: func(ctx context.Context, input *ec2.DeleteSubnetInput, optFns ...func(*ec2.Options)) (*ec2.DeleteSubnetOutput, error) {
						return nil, fmt.Errorf("failed to delete subnet")
					},
					describeInternetGatewaysFunc: func(ctx context.Context, input *ec2.DescribeInternetGatewaysInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInternetGatewaysOutput, error) {
						return &ec2.DescribeInternetGatewaysOutput{
							InternetGateways: []types.InternetGateway{
								{
									InternetGatewayId: aws.String("igw-12345"),
								},
							},
						}, nil
					},
					describeSubnetsFunc: func(ctx context.Context, input *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error) {
						return &ec2.DescribeSubnetsOutput{
							Subnets: []types.Subnet{
								{SubnetId: aws.String("subnet-1")},
								{SubnetId: aws.String("subnet-2")},
							},
						}, nil
					},
					detachInternetGatewayFunc: func(ctx context.Context, input *ec2.DetachInternetGatewayInput, optFns ...func(*ec2.Options)) (*ec2.DetachInternetGatewayOutput, error) {
						return &ec2.DetachInternetGatewayOutput{}, nil
					},
				},
				vpcID: "vpc-12345",
			},
			wantErr: true,
		},
		{
			name: "Error deleting route tables",
			args: args{
				ctx: context.Background(),
				client: &MockEC2Client{
					deleteInternetGatewayFunc: func(ctx context.Context, input *ec2.DeleteInternetGatewayInput, optFns ...func(*ec2.Options)) (*ec2.DeleteInternetGatewayOutput, error) {
						return &ec2.DeleteInternetGatewayOutput{}, nil
					},
					deleteRouteTableFunc: func(ctx context.Context, input *ec2.DeleteRouteTableInput, optFns ...func(*ec2.Options)) (*ec2.DeleteRouteTableOutput, error) {
						return nil, fmt.Errorf("failed to delete route table")
					},
					deleteSubnetFunc: func(ctx context.Context, input *ec2.DeleteSubnetInput, optFns ...func(*ec2.Options)) (*ec2.DeleteSubnetOutput, error) {
						return &ec2.DeleteSubnetOutput{}, nil
					},
					describeInternetGatewaysFunc: func(ctx context.Context, input *ec2.DescribeInternetGatewaysInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInternetGatewaysOutput, error) {
						return &ec2.DescribeInternetGatewaysOutput{
							InternetGateways: []types.InternetGateway{
								{
									InternetGatewayId: aws.String("igw-12345"),
								},
							},
						}, nil
					},
					describeRouteTablesFunc: func(ctx context.Context, input *ec2.DescribeRouteTablesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeRouteTablesOutput, error) {
						return &ec2.DescribeRouteTablesOutput{
							RouteTables: []types.RouteTable{
								{RouteTableId: aws.String("rtb-1")},
								{RouteTableId: aws.String("rtb-2")},
							},
						}, nil
					},
					describeSubnetsFunc: func(ctx context.Context, input *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error) {
						return &ec2.DescribeSubnetsOutput{
							Subnets: []types.Subnet{
								{
									SubnetId: aws.String("subnet-1"),
								},
								{
									SubnetId: aws.String("subnet-2"),
								},
							},
						}, nil
					},
					detachInternetGatewayFunc: func(ctx context.Context, input *ec2.DetachInternetGatewayInput, optFns ...func(*ec2.Options)) (*ec2.DetachInternetGatewayOutput, error) {
						return &ec2.DetachInternetGatewayOutput{}, nil
					},
				},
				vpcID: "vpc-12345",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := cleanupVPCResources(tt.args.ctx, tt.args.client, tt.args.vpcID); (err != nil) != tt.wantErr {
				t.Errorf("cleanupVPCResources() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
