# srcds_watch

Prometheus exporter for srcds stats retrieved via status/stats rcon command.

## Exported Metrics

    # HELP srcds_stats_connects The total number of players that have connected to the server.
    # TYPE srcds_stats_connects gauge

    # HELP srcds_stats_cpu The current cpu usage.
    # TYPE srcds_stats_cpu gauge
    
    # HELP srcds_stats_fps The current server fps (tickrate)
    # TYPE srcds_stats_fps gauge
    
    # HELP srcds_stats_maps The total number of maps that have been played
    # TYPE srcds_stats_maps gauge
    
    # HELP srcds_stats_metamod_version Current currently running metamod version
    # TYPE srcds_stats_metamod_version gauge
    
    # HELP srcds_stats_net_in The current inbound network traffic rate (KB/s)
    # TYPE srcds_stats_net_in gauge
    
    # HELP srcds_stats_net_out The current outbound network traffic rate (KB/s)
    # TYPE srcds_stats_net_out gauge
    
    # HELP srcds_stats_online 1 if the game server is online
    # TYPE srcds_stats_online gauge
    
    # HELP srcds_stats_players The current statusPlayer count of the server.
    # TYPE srcds_stats_players gauge
    
    # HELP srcds_stats_source_tv The current status of source tv
    # TYPE srcds_stats_source_tv gauge

    # HELP srcds_stats_sourcemod_version Current currently running sourcemod version
    # TYPE srcds_stats_sourcemod_version gauge
    
    # HELP srcds_stats_sv_max_update_rate The time in MS per tick
    # TYPE srcds_stats_sv_max_update_rate gauge
    
    # HELP srcds_stats_sv_visiblemaxplayers The currently configured sv_visiblemaxplayers value
    # TYPE srcds_stats_sv_visiblemaxplayers gauge
    
    # HELP srcds_stats_uptime The current server uptime in minutes
    # TYPE srcds_stats_uptime gauge
    
    # HELP srcds_status_connected The duration the player has been connected for in seconds
    # TYPE srcds_status_connected gauge
    
    # HELP srcds_status_edicts The current edict usage (2048 max)
    # TYPE srcds_status_edicts gauge
    
    # HELP srcds_status_loss The current player loss
    # TYPE srcds_status_loss gauge
    
    # HELP srcds_status_ping The current player ping
    # TYPE srcds_status_ping gauge
    
    # HELP srcds_status_players_bots The current server bot player limit
    # TYPE srcds_status_players_bots gauge
    
    # HELP srcds_status_players_count The current server player count
    # TYPE srcds_status_players_count gauge
    
    # HELP srcds_status_players_human The current server human player count
    # TYPE srcds_status_players_human gauge
    
    # HELP srcds_status_players_limit The current server player limit
    # TYPE srcds_status_players_limit gauge


## Docker Example

    docker run -v $(pwd)/srcds_watch.yml:/app/srcds_watch.yml ghcr.io/leighmacdonald/srcds_watch:v1.0.0