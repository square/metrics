### For counters, transform into a rate **before** using any aggregate functions.

Why? Consider two series, `A: [1, 2, 3, 4, 0, 1, 2, 3]` and `B: [1, 2, 3, 4, 5, 6, 7, 8]`. If you transform into a rate first and then aggregate by sum, you will get a consistent rate of 1 for both series, save for a single `NaN` when `A` resets (since MQE can detect the counter reset). On the other hand, if you aggregate by sum first, you will get an incorrect rate when `A` resets.

In short, transforming into a rate first will ensure that your rate is always correct, even when the counter resets.

### To accurately capture peak rates, prefer gauges over counters.

At lower resolutions (i.e. rolled-up data), the rate of a counter is calculated over a longer time period, which means the peak will be lower. For example, let's say we have a series `A: [1, 2, 3, 13, 14, 15]` collected in 30s intervals and the corresponding rolled up series (sampled by max) `A': [1, 15]`. The peak rate for `A` will be `(13 - 3) / 2 / 30 = 0.167` while the peak rate for `A'` will be `(15 - 1) / 2 / 300 = 0.023`.

If you want to truly capture peak values, you should directly emit rates as gauges so that the peaks are preserved when rollups occur. Alternatively, you could store counter rollups as rates so that the loss of resolution does not affect your data.