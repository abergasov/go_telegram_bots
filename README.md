## Logs
```shell script
docker logs -t -f tg_bot.mf
```

## NewBotAdd
1. implement SingleBot interface (pkg/bot_manager/manger.go:14)
1. add bot to botPool in InitBotManager (pkg/bot_manager/manger.go:27)
1. set webHook url from logs

### Sample config
```yaml
host_url: https://bot.site.com
prod_env: true
app_port: 8090
orchestra_url: http://site.com/abc
orchestra_key: ZJzVdtXD56amDYAhuRcFOgUBk6pZJzVdtXD56amDYAhuRcFOgUBk
bot_list:
  - bot_name: "@my_super_bot_1"
    bot_token: 123:abc
    bot_hook_path: 07Q7ZXM2xCUsdvQtpxtaueoeq08PmR0CE4735cr // prefix for webhook updates
    bot_admin_chats:
      - 86003117
    bot_logger_chat:
      - -255525844
    db_conf: // skip this section if no db need
      db_host: 127.0.0.1
      db_user: ryuiJNbLwqtbpG1NYecNbzi
      db_name: TJRDHZ7f
      db_pass: KpU1H43J6bAKRKKjB0IovKrTUTd2AcqbYodS
      db_port: 3821
  - bot_name: "@my_super_bot_2"
    bot_token: 123:abc
    bot_hook_path: WKmIGDeMvWxSCVkwrRgZxkZsUca1ir35ZR0Eez3 // prefix for webhook updates
    bot_admin_chats:
      - 86003117
    bot_logger_chat:
      - -255525844
    db_conf: // skip this section if no db need
      db_host: 127.0.0.1
      db_user: PXnSleOuvaLBNgI2YxJCLRB
      db_name: iFdkVFVl
      db_pass: UcdKCVuqZxwTOtjP3ACLVPPYEil1KTfAbZvO
      db_port: 3519

```