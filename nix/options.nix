{
  config,
  lib,
  pkgs,
  ...
}:
let
  mkHisterEnv =
    cfg:
    lib.optionalAttrs (cfg.dataDir != null) {
      HISTER_DATA_DIR = cfg.dataDir;
    }
    // lib.optionalAttrs (cfg.port != null) {
      HISTER_PORT = builtins.toString cfg.port;
    }
    // lib.optionalAttrs (cfg.configPath != null) {
      HISTER_CONFIG = builtins.toString cfg.configPath;
    }
    // lib.optionalAttrs (cfg.config != null) {
      HISTER_CONFIG = "${(pkgs.formats.yaml { }).generate "hister-config.yml" cfg.config}";
    };
in
{
  options.services.hister = {
    enable = lib.mkEnableOption "Hister web history service";

    package = lib.mkOption {
      type = lib.types.package;
      description = "The hister package to use.";
    };

    dataDir = lib.mkOption {
      type = lib.types.nullOr lib.types.path;
      default = null;
      example = "/var/lib/hister";
      description = ''
        Directory where Hister stores its data.
        If set, this will override the `app.directory` setting in the configuration file.
      '';
    };

    port = lib.mkOption {
      type = lib.types.nullOr lib.types.port;
      default = null;
      example = 4433;
      description = ''
        Port on which Hister listens.
        If set, this will override the `server.address` port in the configuration file.
      '';
    };

    configPath = lib.mkOption {
      type = lib.types.nullOr lib.types.path;
      default = null;
      example = "/etc/hister/config.yml";
      description = "Path to an existing configuration file.";
    };

    config = lib.mkOption {
      type = with lib.types; nullOr attrs;
      default = null;
      description = "Configuration as a Nix attribute set. This will be converted to a YAML file.";
      example = {
        app = {
          search_url = "https://google.com/search?q={query}";
          log_level = "info";
        };
        server = {
          address = "127.0.0.1:4433";
          database = "db.sqlite3";
        };
        hotkeys = {
          "/" = "focus_search_input";
          "enter" = "open_result";
          "alt+enter" = "open_result_in_new_tab";
          "alt+j" = "select_next_result";
          "alt+k" = "select_previous_result";
          "alt+o" = "open_query_in_search_engine";
        };
      };
    };
  };

  config = {
    assertions = [
      {
        assertion = !(config.services.hister.configPath != null && config.services.hister.config != null);
        message = "Only one of services.hister.configPath and services.hister.config can be set";
      }
    ];
    _module.args.histerEnv = mkHisterEnv;
  };
}
