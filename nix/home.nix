{
  config,
  lib,
  histerEnv,
  ...
}:
{
  imports = [
    ./options.nix
  ];

  config = lib.mkIf config.services.hister.enable {
    home.packages = [ config.services.hister.package ];

    systemd.user.services.hister = {
      Unit = {
        Description = "Hister web history service";
        After = [ "network.target" ];
      };

      Service = {
        ExecStart = "${lib.getExe config.services.hister.package} listen";
        Restart = "on-failure";
        WorkingDirectory = lib.mkIf (config.services.hister.dataDir != null) config.services.hister.dataDir;

        Environment = lib.mapAttrsToList (name: value: "${name}=${value}") (
          histerEnv config.services.hister
        );
      };

      Install = {
        WantedBy = [ "default.target" ];
      };
    };

    launchd.agents.hister = {
      enable = true;
      config = {
        ProgramArguments = [
          (lib.getExe config.services.hister.package)
          "listen"
        ];
        KeepAlive = true;
        WorkingDirectory = lib.mkIf (config.services.hister.dataDir != null) config.services.hister.dataDir;

        EnvironmentVariables = histerEnv config.services.hister;
      };
    };
  };
}
