-- name: PriceAt :one
SELECT model, effective_from, input, cache_write_5m, cache_write_1h, cache_read, output
FROM model_prices
WHERE model = ? AND effective_from <= ?
ORDER BY effective_from DESC
LIMIT 1;
