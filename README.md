# httptap


## 

- [ ] TODOs

## Specification schema

kind: TapHandler
spec:
  listenAddress: <string>
  Logger: 
    logLevel: <string debug, info, warn, error>
    logFile: <string> Default stdout
  taps: <[]Tap>
      name: <string>
      patterns: <[]string>
      logTap: <LogTap>
      templateTap: <TemplateTap>
  
LogTap:
  logFile: 