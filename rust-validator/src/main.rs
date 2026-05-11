use std::env;
use std::fs::File;
use std::io::{self, BufRead};
use std::path::Path;
use std::time::Instant;

use lab14_social_validator::validate_record_json;

fn main() {
    let args: Vec<String> = env::args().collect();
    if args.len() < 3 {
        eprintln!("usage: social-validator <validate|bench> <jsonl-path> [iterations]");
        std::process::exit(2);
    }

    let command = &args[1];
    let path = Path::new(&args[2]);
    let result = match command.as_str() {
        "validate" => validate_file(path).map(|count| {
            println!("validated {count} records");
        }),
        "bench" => {
            let iterations = args.get(3).and_then(|value| value.parse::<usize>().ok()).unwrap_or(5);
            benchmark_file(path, iterations)
        }
        _ => Err(format!("unknown command: {command}")),
    };

    if let Err(err) = result {
        eprintln!("{err}");
        std::process::exit(1);
    }
}

fn validate_file(path: &Path) -> Result<usize, String> {
    let file = File::open(path).map_err(|err| format!("open {}: {err}", path.display()))?;
    let mut count = 0;
    for (index, line) in io::BufReader::new(file).lines().enumerate() {
        let line = line.map_err(|err| format!("read line {}: {err}", index + 1))?;
        if line.trim().is_empty() {
            continue;
        }
        validate_record_json(&line).map_err(|err| format!("line {}: {err}", index + 1))?;
        count += 1;
    }
    Ok(count)
}

fn benchmark_file(path: &Path, iterations: usize) -> Result<(), String> {
    let lines = std::fs::read_to_string(path)
        .map_err(|err| format!("read {}: {err}", path.display()))?
        .lines()
        .filter(|line| !line.trim().is_empty())
        .map(ToOwned::to_owned)
        .collect::<Vec<_>>();
    let start = Instant::now();
    for _ in 0..iterations {
        for line in &lines {
            validate_record_json(line)?;
        }
    }
    let elapsed = start.elapsed();
    let total = lines.len() * iterations;
    println!(
        "validated {total} records in {:.3?} ({:.0} records/sec)",
        elapsed,
        total as f64 / elapsed.as_secs_f64()
    );
    Ok(())
}
