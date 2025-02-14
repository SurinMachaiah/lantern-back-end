package config

import "github.com/spf13/viper"

// SetupConfig associates the application with all of the relevant configuration parameters
// for the application with the prefix 'lantern'.
func SetupConfig() error {
	var err error

	viper.SetEnvPrefix("lantern")
	viper.AutomaticEnv()

	// Database setup

	err = viper.BindEnv("dbhost")
	if err != nil {
		return err
	}
	err = viper.BindEnv("dbport")
	if err != nil {
		return err
	}
	err = viper.BindEnv("dbuser")
	if err != nil {
		return err
	}
	err = viper.BindEnv("dbpassword")
	if err != nil {
		return err
	}
	err = viper.BindEnv("dbname")
	if err != nil {
		return err
	}
	err = viper.BindEnv("dbsslmode")
	if err != nil {
		return err
	}
	err = viper.BindEnv("chplapikey")
	if err != nil {
		return err
	}

	// Capability Queue Setup

	err = viper.BindEnv("quser")
	if err != nil {
		return err
	}
	err = viper.BindEnv("qpassword")
	if err != nil {
		return err
	}
	err = viper.BindEnv("qhost")
	if err != nil {
		return err
	}
	err = viper.BindEnv("qport")
	if err != nil {
		return err
	}
	err = viper.BindEnv("capquery_qryintvl") // in minutes
	if err != nil {
		return err
	}

	// Version Response Queue Setup
	err = viper.BindEnv("versionsquery_qname")
	if err != nil {
		return err
	}
	err = viper.BindEnv("versionsquery_response_qname")
	if err != nil {
		return err
	}

	// Version Response Queue Setup
	err = viper.BindEnv("versionsquery_qname")
	if err != nil {
		return err
	}
	err = viper.BindEnv("versionsquery_response_qname")
	if err != nil {
		return err
	}

	// Info History Pruning
	err = viper.BindEnv("pruning_threshold") // in minutes
	if err != nil {
		return err
	}

	err = viper.BindEnv("export_numworkers")
	if err != nil {
		return err
	}
	err = viper.BindEnv("export_duration")
	if err != nil {
		return err
	}

	viper.SetDefault("dbhost", "localhost")
	viper.SetDefault("dbport", 5432)
	viper.SetDefault("dbuser", "lantern")
	viper.SetDefault("dbpassword", "postgrespassword")
	viper.SetDefault("dbname", "lantern")
	viper.SetDefault("dbsslmode", "disable")

	viper.SetDefault("quser", "capabilityquerier")
	viper.SetDefault("qpassword", "capabilityquerier")
	viper.SetDefault("qhost", "localhost")
	viper.SetDefault("qport", "5672")
	viper.SetDefault("capquery_qname", "capability-statements")
	viper.SetDefault("endptinfo_capquery_qname", "endpoints-to-capability")
	viper.SetDefault("versionsquery_qname", "version-responses")
	viper.SetDefault("versionsquery_response_qname", "endpoints-to-version-responses")
	viper.SetDefault("capquery_qryintvl", 1380) // 1380 minutes -> 23 hours.

	viper.SetDefault("pruning_threshold", 43800) // 43800 minutes -> 1 month.

	viper.SetDefault("export_numworkers", 25)
	viper.SetDefault("export_duration", 240)

	return nil
}

// SetupConfigForTests associates the application with all of the relevant configuration parameters
// for the application and replaces the prefix 'lantern' with 'lantern_test' for the following
// environment variables:
// - dbuser
// - dbpassword
// - dbname
func SetupConfigForTests() error {
	var err error

	err = SetupConfig()
	if err != nil {
		return err
	}

	prevDbName := viper.GetString("dbname")

	viper.SetEnvPrefix("lantern_test")
	viper.AutomaticEnv()

	err = viper.BindEnv("dbuser")
	if err != nil {
		return err
	}
	err = viper.BindEnv("dbpassword")
	if err != nil {
		return err
	}
	err = viper.BindEnv("dbname")
	if err != nil {
		return err
	}

	viper.SetDefault("dbuser", "lantern")
	viper.SetDefault("dbpassword", "postgrespassword")
	viper.SetDefault("dbname", "lantern_test")

	if prevDbName == viper.GetString("dbname") {
		panic("Test database and dev/prod database must be different. Test database: " + viper.GetString("dbname") + ". Prod/Dev dataabse: " + prevDbName)
	}

	prevQName := viper.GetString("capquery_qname")

	viper.SetEnvPrefix("lantern_test")
	viper.AutomaticEnv()

	err = viper.BindEnv("quser")
	if err != nil {
		return err
	}
	err = viper.BindEnv("qpassword")
	if err != nil {
		return err
	}

	// Version Response Queue Setup
	err = viper.BindEnv("versionsquery_qname")
	if err != nil {
		return err
	}
	err = viper.BindEnv("versionsquery_response_qname")
	if err != nil {
		return err
	}

	// Version Response Queue Setup
	err = viper.BindEnv("versionsquery_qname")
	if err != nil {
		return err
	}
	err = viper.BindEnv("versionsquery_response_qname")
	if err != nil {
		return err
	}

	viper.SetDefault("quser", "capabilityquerier")
	viper.SetDefault("qpassword", "capabilityquerier")
	viper.SetDefault("qname", "test-queue")
	viper.SetDefault("endptinfo_capquery_qname", "test-endpoints-to-capability")
	viper.SetDefault("versionsquery_qname", "test-version-responses")
	viper.SetDefault("versionsquery_response_qname", "test-endpoints-to-version-responses")

	if prevQName == viper.GetString("qname") {
		panic("Test queue and dev/prod queue must be different. Test queue: " + viper.GetString("qname") + ". Prod/Dev queue: " + prevQName)
	}

	return nil
}
