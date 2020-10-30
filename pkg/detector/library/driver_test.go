package library_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aquasecurity/trivy-db/pkg/db"
	"github.com/aquasecurity/trivy/pkg/detector/library"
	"github.com/aquasecurity/trivy/pkg/types"
	"github.com/aquasecurity/trivy/pkg/utils"
)

func TestDriver_Detect(t *testing.T) {
	type fields struct {
		fileName string
	}
	type args struct {
		pkgName string
		pkgVer  string
	}
	tests := []struct {
		name     string
		fixtures []string
		fields   fields
		args     args
		want     []types.DetectedVulnerability
		wantErr  string
	}{
		{
			name:     "happy path",
			fixtures: []string{"testdata/fixtures/php.yaml"},
			fields:   fields{fileName: "composer.lock"},
			args: args{
				pkgName: "symfony/symfony",
				pkgVer:  "4.2.6",
			},
			want: []types.DetectedVulnerability{
				{
					VulnerabilityID:  "CVE-2019-10909",
					PkgName:          "symfony/symfony",
					InstalledVersion: "4.2.6",
					FixedVersion:     "4.2.7",
					URL:              "https://avd.aquasec.com/nvd/cve-2019-10909",
				},
			},
		},
		{
			name:     "happy path, python with custom vulnerability ID",
			fixtures: []string{"testdata/fixtures/python.yaml"},
			fields:   fields{fileName: "Pipfile.lock"},
			args: args{
				pkgName: "django-cors-headers/django-cors-headers",
				pkgVer:  semver.MustParse("2.5.2"),
			},
			want: []types.DetectedVulnerability{
				{
					VulnerabilityID:  "pyup.io-37132",
					PkgName:          "django-cors-headers/django-cors-headers",
					InstalledVersion: "2.5.2",
					FixedVersion:     ">=3.0.0",
				},
			},
		},
		{
			name:     "non-prefix buckets",
			fixtures: []string{"testdata/fixtures/php-without-prefix.yaml"},
			fields:   fields{fileName: "composer.lock"},
			args: args{
				pkgName: "symfony/symfony",
				pkgVer:  "4.2.6",
			},
			want: []types.DetectedVulnerability{
				{
					VulnerabilityID:  "CVE-2019-10909",
					PkgName:          "symfony/symfony",
					InstalledVersion: "4.2.6",
					FixedVersion:     "4.2.7",
				},
			},
		},
		{
			name:     "no patched versions in the advisory",
			fixtures: []string{"testdata/fixtures/php.yaml"},
			fields:   fields{fileName: "composer.lock"},
			args: args{
				pkgName: "symfony/symfony",
				pkgVer:  "4.4.6",
			},
			want: []types.DetectedVulnerability{
				{
					VulnerabilityID:  "CVE-2020-5275",
					PkgName:          "symfony/symfony",
					InstalledVersion: "4.4.6",
					FixedVersion:     "4.4.7",
					URL:              "https://avd.aquasec.com/nvd/cve-2020-5275",
				},
			},
		},
		{
			name:     "no vulnerable versions in the advisory",
			fixtures: []string{"testdata/fixtures/ruby.yaml"},
			fields:   fields{fileName: "Gemfile.lock"},
			args: args{
				pkgName: "activesupport",
				pkgVer:  "4.1.1",
			},
			want: []types.DetectedVulnerability{
				{
					VulnerabilityID:  "CVE-2015-3226",
					PkgName:          "activesupport",
					InstalledVersion: "4.1.1",
					FixedVersion:     ">= 4.2.2, ~> 4.1.11",
					URL:              "https://avd.aquasec.com/nvd/cve-2015-3226",
				},
			},
		},
		{
			name:     "no vulnerability",
			fixtures: []string{"testdata/fixtures/php.yaml"},
			fields:   fields{fileName: "composer.lock"},
			args: args{
				pkgName: "symfony/symfony",
				pkgVer:  "4.4.7",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize DB
			dir := utils.InitTestDB(t, tt.fixtures)
			defer os.RemoveAll(dir)
			defer db.Close()

			factory := library.DriverFactory{}
			driver, err := factory.NewDriver(tt.fields.fileName)
			require.NoError(t, err)

			got, err := driver.Detect(tt.args.pkgName, tt.args.pkgVer)
			switch {
			case tt.wantErr != "":
				require.NotNil(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			default:
				assert.NoError(t, err)
			}

			// Compare
			assert.Equal(t, tt.want, got)
		})
	}
}
