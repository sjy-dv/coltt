// build.rs
fn main() -> Result<(), Box<dyn std::error::Error>> {
    tonic_build::configure()
        .build_server(true)
        // .out_dir("gen")
        .compile_protos(
            &["proto/edge.proto"],
            &["proto"],
        )?;
    Ok(())
}
