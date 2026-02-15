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

  options.services.hister = {
    user = lib.mkOption {
      type = lib.types.str;
      default = "hister";
      description = "User account under which Hister runs.";
    };

    group = lib.mkOption {
      type = lib.types.str;
      default = "hister";
      description = "Group under which Hister runs.";
    };
  };

  config = lib.mkIf config.services.hister.enable {
    environment.systemPackages = [ config.services.hister.package ];

    users.users = lib.mkIf (config.services.hister.user == "hister") {
      hister = {
        description = "Hister web history service";
        group = config.services.hister.group;
        isSystemUser = true;
      };
    };

    users.groups = lib.mkIf (config.services.hister.group == "hister") {
      hister = { };
    };

    systemd.services.hister = {
      description = "Hister web history service";
      after = [ "network.target" ];
      wantedBy = [ "multi-user.target" ];

      environment = histerEnv config.services.hister;

      serviceConfig = {
        ExecStart = "${lib.getExe config.services.hister.package} listen";
        Restart = "on-failure";
        User = config.services.hister.user;
        Group = config.services.hister.group;
        StateDirectory = "hister";
      };
    };

    networking.firewall.allowedTCPPorts = lib.mkIf (config.services.hister.port != null) [
      config.services.hister.port
    ];
  };
}
