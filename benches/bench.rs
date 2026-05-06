//! Benchmark module placeholder.

use criterion::{black_box, criterion_group, criterion_main, Criterion};
use sha2::{Digest, Sha256};

fn bench_hash(c: &mut Criterion) {
    let data = b"hello world";
    c.bench_function("sha256_hash", |b| {
        b.iter(|| {
            let mut hasher = Sha256::new();
            hasher.update(black_box(data));
            hasher.finalize()
        });
    });
}

criterion_group!(benches, bench_hash);
criterion_main!(benches);