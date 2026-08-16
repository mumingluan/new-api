package limiter

import (
	"context"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type commandCounterHook struct {
	eval    atomic.Int64
	evalSha atomic.Int64
}

func (h *commandCounterHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	switch cmd.Name() {
	case "eval":
		h.eval.Add(1)
	case "evalsha":
		h.evalSha.Add(1)
	}
	return ctx, nil
}

func (h *commandCounterHook) AfterProcess(context.Context, redis.Cmder) error {
	return nil
}

func (h *commandCounterHook) BeforeProcessPipeline(ctx context.Context, _ []redis.Cmder) (context.Context, error) {
	return ctx, nil
}

func (h *commandCounterHook) AfterProcessPipeline(context.Context, []redis.Cmder) error {
	return nil
}

func TestAllowLoadsMissingScriptAndRecoversAfterFlush(t *testing.T) {
	ctx := context.Background()
	redisServer := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	limiter := New(ctx, client)

	exists, err := rateLimitScript.Exists(ctx, client).Result()
	require.NoError(t, err)
	require.Equal(t, []bool{false}, exists)

	allowed, err := limiter.Allow(ctx, "initial")
	require.NoError(t, err)
	assert.True(t, allowed)

	exists, err = rateLimitScript.Exists(ctx, client).Result()
	require.NoError(t, err)
	require.Equal(t, []bool{true}, exists)
	require.NoError(t, client.ScriptFlush(ctx).Err())

	allowed, err = limiter.Allow(ctx, "after-flush")
	require.NoError(t, err)
	assert.True(t, allowed)
	exists, err = rateLimitScript.Exists(ctx, client).Result()
	require.NoError(t, err)
	assert.Equal(t, []bool{true}, exists)
}

func TestAllowConcurrentRecoveryAfterScriptFlush(t *testing.T) {
	ctx := context.Background()
	redisServer := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	limiter := New(ctx, client)

	allowed, err := limiter.Allow(ctx, "prime")
	require.NoError(t, err)
	require.True(t, allowed)
	require.NoError(t, client.ScriptFlush(ctx).Err())

	const workers = 16
	var waitGroup sync.WaitGroup
	errors := make(chan error, workers)
	results := make(chan bool, workers)
	for i := 0; i < workers; i++ {
		waitGroup.Add(1)
		go func(worker int) {
			defer waitGroup.Done()
			allowed, err := limiter.Allow(ctx, "concurrent:"+strconv.Itoa(worker))
			errors <- err
			results <- allowed
		}(i)
	}
	waitGroup.Wait()
	close(errors)
	close(results)

	for err := range errors {
		assert.NoError(t, err)
	}
	for allowed := range results {
		assert.True(t, allowed)
	}
}

func TestAllowDoesNotRetryNonNoScriptError(t *testing.T) {
	ctx := context.Background()
	redisServer := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	counter := &commandCounterHook{}
	client.AddHook(counter)
	redisServer.SetError("LOADING Redis is loading the dataset in memory")

	allowed, err := New(ctx, client).Allow(ctx, "loading")

	assert.False(t, allowed)
	require.Error(t, err)
	assert.ErrorContains(t, err, "LOADING Redis is loading the dataset in memory")
	assert.Equal(t, int64(1), counter.evalSha.Load())
	assert.Equal(t, int64(0), counter.eval.Load())
}
