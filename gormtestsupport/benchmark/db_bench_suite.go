package benchmark

import (
	"os"

	config "github.com/fabric8-services/fabric8-common/configuration"
	"github.com/fabric8-services/fabric8-common/gormsupport/cleaner"
	"github.com/fabric8-services/fabric8-common/log"
	"github.com/fabric8-services/fabric8-common/migration"
	"github.com/fabric8-services/fabric8-common/models"
	"github.com/fabric8-services/fabric8-common/resource"
	"github.com/fabric8-services/fabric8-common/test"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq" // need to import postgres driver
	"golang.org/x/net/context"
)

var _ test.SetupAllSuite = &DBBenchSuite{}
var _ test.TearDownAllSuite = &DBBenchSuite{}

// NewDBBenchSuite instanciate a new DBBenchSuite
func NewDBBenchSuite(configFilePath string) DBBenchSuite {
	return DBBenchSuite{configFile: configFilePath}
}

// DBBenchSuite is a base for tests using a gorm db
type DBBenchSuite struct {
	test.Suite
	configFile    string
	Configuration *config.Registry
	DB            *gorm.DB
	Ctx           context.Context
	clean         func()
}

// SetupSuite implements suite.SetupAllSuite
func (s *DBBenchSuite) SetupSuite() {
	resource.Require(s.B(), resource.Database)
	configuration, err := config.New(s.configFile)
	if err != nil {
		log.Panic(nil, map[string]interface{}{
			"err": err,
		}, "failed to setup the configuration")
	}
	s.Configuration = configuration
	if _, c := os.LookupEnv(resource.Database); c != false {
		s.DB, err = gorm.Open("postgres", s.Configuration.GetPostgresConfigString())
		if err != nil {
			log.Panic(nil, map[string]interface{}{
				"err":             err,
				"postgres_config": configuration.GetPostgresConfigString(),
			}, "failed to connect to the database")
		}
	}
	s.Ctx = migration.NewMigrationContext(context.Background())
	s.populateDBBenchSuite(s.Ctx)
}

// populateDBBenchSuite populates the DB with common values
func (s *DBBenchSuite) populateDBBenchSuite(ctx context.Context) {
	if _, c := os.LookupEnv(resource.Database); c != false {
		if err := models.Transactional(s.DB, func(tx *gorm.DB) error {
			return migration.PopulateCommonTypes(ctx, tx)
		}); err != nil {
			log.Panic(nil, map[string]interface{}{
				"err":             err,
				"postgres_config": s.Configuration.GetPostgresConfigString(),
			}, "failed to populate the database with common types")
		}
	}
}

// TearDownSuite implements suite.TearDownAllSuite
func (s *DBBenchSuite) TearDownSuite() {
	s.DB.Close()
}

func (s *DBBenchSuite) SetupBenchmark() {
	s.clean = cleaner.DeleteCreatedEntities(s.DB)
}

func (s *DBBenchSuite) TearDownBenchmark() {
	s.clean()
}
