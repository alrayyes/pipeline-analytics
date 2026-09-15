package sqlite_test

import (
	"context"
	"testing"
	"time"

	"github.com/alrayyes/pipeline-analytics/internal/auth"
	"github.com/alrayyes/pipeline-analytics/internal/auth/sqlite"
	"github.com/alrayyes/pipeline-analytics/internal/db"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/stretchr/testify/require"
)

func newTestStore(t *testing.T) *sqlite.Store {
	t.Helper()

	conn, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, conn.Close()) })
	require.NoError(t, db.Migrate(context.Background(), conn))

	return sqlite.NewStore(conn)
}

func TestStore_CreateUser(t *testing.T) {
	t.Parallel()

	t.Run("creates the first user", func(t *testing.T) {
		t.Parallel()

		store := newTestStore(t)
		user, err := store.CreateUser(context.Background(), []byte("handle-1"), "admin")
		require.NoError(t, err)
		require.NotEmpty(t, user.ID)
		require.Equal(t, "admin", user.DisplayName)
	})

	t.Run("rejects a second user", func(t *testing.T) {
		t.Parallel()

		store := newTestStore(t)
		ctx := context.Background()
		_, err := store.CreateUser(ctx, []byte("handle-1"), "admin")
		require.NoError(t, err)

		_, err = store.CreateUser(ctx, []byte("handle-2"), "someone-else")
		require.ErrorIs(t, err, auth.ErrUserExists)
	})
}

func TestStore_GetUser(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrNoUser before any registration", func(t *testing.T) {
		t.Parallel()

		store := newTestStore(t)
		_, err := store.GetUser(context.Background())
		require.ErrorIs(t, err, auth.ErrNoUser)
	})

	t.Run("returns the created user", func(t *testing.T) {
		t.Parallel()

		store := newTestStore(t)
		ctx := context.Background()
		created, err := store.CreateUser(ctx, []byte("handle-1"), "admin")
		require.NoError(t, err)

		got, err := store.GetUser(ctx)
		require.NoError(t, err)
		require.Equal(t, created.ID, got.ID)
		require.Equal(t, []byte("handle-1"), got.UserHandle)
	})
}

func TestStore_Credentials(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	ctx := context.Background()
	user, err := store.CreateUser(ctx, []byte("handle-1"), "admin")
	require.NoError(t, err)

	cred := webauthn.Credential{ID: []byte("cred-1"), PublicKey: []byte("pubkey"), Flags: webauthn.CredentialFlags{UserPresent: true}}
	require.NoError(t, store.PutCredential(ctx, user.ID, cred))

	t.Run("round trips a stored credential", func(t *testing.T) {
		t.Parallel()

		creds, err := store.CredentialsForUser(ctx, user.ID)
		require.NoError(t, err)
		require.Len(t, creds, 1)
		require.Equal(t, cred.ID, creds[0].ID)
		require.Equal(t, cred.PublicKey, creds[0].PublicKey)
	})

	t.Run("updates in place on a second PutCredential for the same id", func(t *testing.T) {
		t.Parallel()

		updated := cred
		updated.Authenticator.SignCount = 7

		require.NoError(t, store.PutCredential(ctx, user.ID, updated))

		creds, err := store.CredentialsForUser(ctx, user.ID)
		require.NoError(t, err)
		require.Len(t, creds, 1)
		require.Equal(t, uint32(7), creds[0].Authenticator.SignCount)
	})
}

func TestStore_Ceremony(t *testing.T) {
	t.Parallel()

	t.Run("round trips and then deletes", func(t *testing.T) {
		t.Parallel()

		store := newTestStore(t)
		ctx := context.Background()
		require.NoError(t, store.SaveCeremony(ctx, "ceremony-1", []byte("data"), time.Now().Add(time.Minute)))

		got, err := store.TakeCeremony(ctx, "ceremony-1")
		require.NoError(t, err)
		require.Equal(t, []byte("data"), got)

		_, err = store.TakeCeremony(ctx, "ceremony-1")
		require.ErrorIs(t, err, auth.ErrCeremonyNotFound)
	})

	t.Run("an expired ceremony is not returned", func(t *testing.T) {
		t.Parallel()

		store := newTestStore(t)
		ctx := context.Background()
		require.NoError(t, store.SaveCeremony(ctx, "ceremony-2", []byte("data"), time.Now().Add(-time.Minute)))

		_, err := store.TakeCeremony(ctx, "ceremony-2")
		require.ErrorIs(t, err, auth.ErrCeremonyNotFound)
	})
}

func TestStore_Session(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	ctx := context.Background()
	user, err := store.CreateUser(ctx, []byte("handle-1"), "admin")
	require.NoError(t, err)

	t.Run("round trips and then deletes", func(t *testing.T) {
		t.Parallel()

		id, err := store.CreateSession(ctx, user.ID)
		require.NoError(t, err)

		gotUserID, err := store.Session(ctx, id)
		require.NoError(t, err)
		require.Equal(t, user.ID, gotUserID)

		require.NoError(t, store.DeleteSession(ctx, id))

		_, err = store.Session(ctx, id)
		require.ErrorIs(t, err, auth.ErrSessionNotFound)
	})
}
