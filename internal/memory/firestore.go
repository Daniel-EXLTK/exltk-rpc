package memory

import (
	"context"
	"fmt"
	"os"
	"time"

	cloudfirestore "cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

type FirestoreMemory struct {
	client     *cloudfirestore.Client
	collection string
}

type SessionContext struct {
	SessionID string                 `firestore:"session_id" json:"session_id"`
	UserID    string                 `firestore:"user_id" json:"user_id"`
	Context   map[string]interface{} `firestore:"context" json:"context"`
	UpdatedAt time.Time              `firestore:"updated_at" json:"updated_at"`
}

func NewFirestoreMemory(ctx context.Context) (*FirestoreMemory, error) {
	projectID := os.Getenv("FIRESTORE_PROJECT_ID")
	collection := os.Getenv("FIRESTORE_COLLECTION")
	creds := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if projectID == "" || collection == "" || creds == "" {
		return nil, fmt.Errorf("FIRESTORE_PROJECT_ID, FIRESTORE_COLLECTION y GOOGLE_APPLICATION_CREDENTIALS deben estar definidos")
	}
	client, err := cloudfirestore.NewClient(ctx, projectID, option.WithCredentialsFile(creds))
	if err != nil {
		return nil, err
	}
	return &FirestoreMemory{client: client, collection: collection}, nil
}

func (fm *FirestoreMemory) SaveSessionContext(ctx context.Context, session SessionContext) error {
	_, err := fm.client.Collection(fm.collection).Doc(session.SessionID).Set(ctx, session)
	return err
}

func (fm *FirestoreMemory) GetSessionContext(ctx context.Context, sessionID string) (*SessionContext, error) {
	dsnap, err := fm.client.Collection(fm.collection).Doc(sessionID).Get(ctx)
	if err != nil {
		return nil, err
	}
	var session SessionContext
	err = dsnap.DataTo(&session)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (fm *FirestoreMemory) DeleteSessionContext(ctx context.Context, sessionID string) error {
	_, err := fm.client.Collection(fm.collection).Doc(sessionID).Delete(ctx)
	return err
}
