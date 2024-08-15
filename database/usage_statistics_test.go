package database

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestCrudStatistics insert, update & select statistics.
func TestCrudStatistics(t *testing.T) {
	user := User{ID: uuid.NullUUID{}, ExtID: uuid.NewString(), Name: "John Doe", Email: "johndoe@example.org", ProfilePicture: "pic.jpg"}
	user.ID.UUID = uuid.New()
	user.ID.Valid = true
	db := GetConnection()
	_, err := db.NamedExec(SqlUserInsert, user)
	if err != nil {
		t.Fatalf("Error while inserting into database: %+v", err)
	}

	var createdUser User
	err = db.Get(&createdUser, SqlUserSelect, user.ID)
	if err != nil {
		t.Fatalf("Error while reading from database: %+v", err)
	}

	usageStatistics := UsageStatistics{ID: uuid.New(), UserId: createdUser.ID.UUID, Date: time.Now(), TokensIn: 100, TokensOut: 200}
	_, err = db.NamedExec(SqlUsageStatisticsInsert, usageStatistics)
	if err != nil {
		t.Fatalf("Error while inserting into database: %+v", err)
	}

	var createdUsageStatistics UsageStatistics
	err = db.Get(&createdUsageStatistics, SqlUsageStatisticsSelect, usageStatistics.ID)
	if err != nil {
		t.Fatalf("Error while reading from database: %+v", err)
	}
	createdUsageStatistics.Date = usageStatistics.Date
	if createdUsageStatistics != usageStatistics {
		t.Fatalf("Data was not saved correctly: %+v, %+v", createdUsageStatistics, usageStatistics)
	}

	usageStatistics.TokensIn += 555
	usageStatistics.TokensOut += 666

	db.NamedExec(SqlUsageStatisticsUpdate, usageStatistics)
	var updatedUsageStatistics UsageStatistics
	db.Get(&updatedUsageStatistics, SqlUsageStatisticsSelect, usageStatistics.ID)
	updatedUsageStatistics.Date = usageStatistics.Date
	if updatedUsageStatistics != usageStatistics {
		t.Fatalf("Data was not updated correctly: %+v, %+v", updatedUsageStatistics, usageStatistics)
	}

	var UsageStatisticss []UsageStatistics
	db.Select(&UsageStatisticss, SqlUsageStatisticsSelectByUser, user.ID)

	if len(UsageStatisticss) == 0 {
		t.Fatalf("UsageStatisticss for user %+v were not retrieved", user.ID)
	}

	db.Select(&UsageStatisticss, SqlUsageStatisticsSelectByUser, user.ID)

	if len(UsageStatisticss) == 0 {
		t.Fatalf("UsageStatisticss for user %+v were not retrieved", user.ID)
	}
}
