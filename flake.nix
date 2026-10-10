{
  description = "qwer";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/master";
  };

  outputs =
    {
      self,
      nixpkgs,
    }:
    let
      # nix system -> zig target; dist/qwer-<system> is built by ./dist.sh.
      zigTargets = {
        x86_64-linux = "x86_64-linux-musl";
        aarch64-darwin = "aarch64-macos";
      };
      forAllSystems =
        f:
        nixpkgs.lib.genAttrs (builtins.attrNames zigTargets) (
          system:
          f {
            inherit system;
            pkgs = import nixpkgs { inherit system; };
          }
        );
      version = "0.1.0";
    in
    {
      packages = forAllSystems (
        {
          system,
          pkgs,
          ...
        }:
        let
          # Zig package dir for `zig build --system`; entries are named by the
          # hashes in build.zig.zon.
          zigDeps = pkgs.linkFarm "qwer-zig-deps" [
            {
              name = "N-V-__8AAOdBEgAyOZrfIS_QwIGQaghYMBYfIPhZb5j3fANV";
              path = pkgs.fetchFromGitHub {
                owner = "mity";
                repo = "md4c";
                rev = "d61477642dab7cf9efaa30b0b1795a9dc3f02fb0";
                hash = "sha256-xfm72/zXQvc2Skjh1YTn8flyqESY4OugIUcomJvy3fY=";
              };
            }
          ];
        in
        {
          # Prebuilt binary committed in dist/.
          default = pkgs.runCommand "qwer-${version}" { meta.mainProgram = "qwer"; } ''
            install -Dm755 ${./dist + "/qwer-${system}"} $out/bin/qwer
          '';

          # Same build as dist.sh, from source.
          source = pkgs.stdenvNoCC.mkDerivation {
            pname = "qwer";
            inherit version;
            src = pkgs.lib.fileset.toSource {
              root = ./.;
              fileset = pkgs.lib.fileset.unions [
                ./build.zig
                ./build.zig.zon
                ./src
              ];
            };
            nativeBuildInputs = [ pkgs.zig_0_17.hook ];
            dontSetZigDefaultFlags = true;
            zigBuildFlags = [
              "-Dcpu=baseline"
              "-Doptimize=ReleaseSmall"
              "-Dtarget=${zigTargets.${system}}"
              "--system"
              "${zigDeps}"
            ];
            meta.mainProgram = "qwer";
          };
        }
      );

      # `nix develop` provides a shell containing development tools.
      devShell = forAllSystems (
        {
          system,
          pkgs,
        }:
        pkgs.mkShell {
          buildInputs = [ pkgs.zig_0_17 ];
        }
      );

      overlays.default = final: prev: {
        # diff2 = self.packages.${final.stdenv.system}.diff2;
      };
    };
}
