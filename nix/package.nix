{
  lib,
  buildGoModule,
  histerRev ? "unknown",
}:
let
  packageJson = builtins.fromJSON (builtins.readFile ../ext/package.json);
in
buildGoModule (finalAttrs: {
  pname = "hister";
  version = packageJson.version;

  src = ../.;

  vendorHash = "sha256-9Pg3NMZJjrKJjMXPWnN37ivNno/L14n7r/cbMhLqIjU=";

  ldflags = [
    "-s"
    "-w"
    "-X main.version=${finalAttrs.version}"
    "-X main.commit=${histerRev}"
  ];

  subPackages = [ "." ];

  doCheck = false;

  meta = {
    description = "Web history on steroids - blazing fast, content-based search for visited websites";
    homepage = "https://github.com/asciimoo/hister";
    license = lib.licenses.agpl3Only;
    maintainers = [ lib.maintainers.FlameFlag ];
    mainProgram = "hister";
    platforms = lib.platforms.unix;
  };
})
