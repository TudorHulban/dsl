package main

// --- Start of DSL for the current dataset ---

// criteria "criteria_name_1" {
// 	// Optional settings (baselines, increments)
// 	baseline setting_name = value;

// 	monitor "column_name_a" {
// 	  level n when condition;
// 	  // ... more rules for column_a
// 	}

// 	monitor "column_name_b" {
// 	  level m when condition;
// 	  // ... more rules for column_b
// 	}
// 	// ... more monitors
//   } // end criteria_name_1

//   criteria "criteria_name_2" {
// 	monitor "column_name_c" {
// 	   level p when condition;
// 	}
// 	// ... more monitors or settings
//   } // end criteria_name_2

//   // ... potentially more criteria blocks

//   // --- End of DSL ---

const (
	_dslCriteria  = "criteria"
	_dslThreshold = "threshold"
)
