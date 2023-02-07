package rule

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws-controllers-k8s/eventbridge-controller/apis/v1alpha1"
)

func Test_validateRuleSpec(t *testing.T) {
	type args struct {
		spec v1alpha1.RuleSpec
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "invalid state",
			args: args{
				spec: v1alpha1.RuleSpec{
					State: aws.String("invalid"),
				},
			},
			wantErr: true,
		},
		{
			name: "invalid state (empty string)",
			args: args{
				spec: v1alpha1.RuleSpec{
					State: aws.String(""),
				},
			},
			wantErr: true,
		},
		{
			name: "invalid target (missing arn)",
			args: args{
				spec: v1alpha1.RuleSpec{
					State:        aws.String("ENABLED"),
					EventPattern: aws.String(`{"some":"pattern"}`),
					Targets: []*v1alpha1.Target{
						{
							ARN: nil,
							ID:  aws.String("some-id"),
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid target (missing id)",
			args: args{
				spec: v1alpha1.RuleSpec{
					State:        aws.String("ENABLED"),
					EventPattern: aws.String(`{"some":"pattern"}`),
					Targets: []*v1alpha1.Target{
						{
							ARN: aws.String("some-arn"),
							ID:  nil,
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid state (lower case)",
			args: args{
				spec: v1alpha1.RuleSpec{
					State: aws.String("enabled"),
				},
			},
			wantErr: true,
		},
		{
			name: "valid state, pattern missing",
			args: args{
				spec: v1alpha1.RuleSpec{
					State: aws.String("ENABLED"),
				},
			},
			wantErr: true,
		},
		{
			name: "valid state, rule and schedule pattern specified",
			args: args{
				spec: v1alpha1.RuleSpec{
					State:              aws.String("ENABLED"),
					EventPattern:       aws.String(`{"some":"pattern"}`),
					ScheduleExpression: aws.String(`{"someschedule}`),
				},
			},
			wantErr: false,
		},
		{
			name: "valid state and rule pattern",
			args: args{
				spec: v1alpha1.RuleSpec{
					State:        aws.String("ENABLED"),
					EventPattern: aws.String(`{"some":"pattern"}`), // we don't verify rule syntax
				},
			},
			wantErr: false,
		},
		{
			name: "valid state and schedule pattern",
			args: args{
				spec: v1alpha1.RuleSpec{
					State:              aws.String("ENABLED"),
					ScheduleExpression: aws.String(`{"someschedule"}`), // we don't verify rule syntax
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateRuleSpec(tt.args.spec); (err != nil) != tt.wantErr {
				t.Errorf("validateRuleSpec() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
