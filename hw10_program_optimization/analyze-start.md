
##  Было

```bash
$ go test -bench=BenchmarkGetDomainStat -benchtime=10s -cpuprofile=cpu.prof -memprofile=mem.prof
```

```bash 
goos: linux
goarch: amd64
pkg: github.com/fixme_my_friend/hw10_program_optimization
cpu: AMD Ryzen 9 9950X 16-Core Processor
BenchmarkGetDomainStat-32            184          65794745 ns/op
PASS
ok      github.com/fixme_my_friend/hw10_program_optimization    18.958s
```

```bash
$ go tool pprof -text -nodecount=20 -alloc_space mem.prof
```

```bash
File: hw10_program_optimization.test
Build ID: 1abefdb2009a1e6b4e482c1b5829f1dc53029a9b
Type: alloc_space
Time: 2026-03-18 22:01:41 MSK
Showing nodes accounting for 41.99GB, 100% of 42.01GB total
Dropped 53 nodes (cum <= 0.21GB)
Showing top 20 nodes out of 24
      flat  flat%   sum%        cum   cum%
   16.29GB 38.78% 38.78%    16.29GB 38.78%  regexp/syntax.(*compiler).inst (inline)
    8.94GB 21.29% 60.07%     8.94GB 21.29%  regexp/syntax.(*parser).newRegexp (inline)
    5.58GB 13.29% 73.36%       42GB   100%  github.com/fixme_my_friend/hw10_program_optimization.GetDomainStat
    4.54GB 10.80% 84.15%    36.17GB 86.11%  regexp.compile
    3.41GB  8.11% 92.27%    13.44GB 31.99%  regexp/syntax.parse
    1.26GB  2.99% 95.26%     2.54GB  6.04%  regexp/syntax.(*compiler).init (inline)
    0.66GB  1.57% 96.83%     1.09GB  2.59%  regexp/syntax.(*parser).push
    0.43GB  1.02% 97.85%     0.43GB  1.02%  regexp/syntax.(*parser).maybeConcat
    0.42GB  1.01% 98.85%     0.42GB  1.01%  regexp/syntax.(*Regexp).CapNames
    0.23GB  0.56% 99.41%    36.42GB 86.69%  github.com/fixme_my_friend/hw10_program_optimization.countDomains
    0.23GB  0.54%   100%     0.23GB  0.54%  unicode/utf8.AppendRune (inline)
         0     0%   100%    41.51GB 98.82%  github.com/fixme_my_friend/hw10_program_optimization.BenchmarkGetDomainStat     
         0     0%   100%    36.17GB 86.11%  regexp.Compile (inline)
         0     0%   100%    36.18GB 86.13%  regexp.Match
         0     0%   100%     0.23GB  0.54%  regexp/syntax.(*Prog).Prefix
         0     0%   100%    15.01GB 35.73%  regexp/syntax.(*compiler).compile
         0     0%   100%    15.01GB 35.73%  regexp/syntax.(*compiler).rune
         0     0%   100%     7.08GB 16.85%  regexp/syntax.(*parser).literal
         0     0%   100%    17.55GB 41.77%  regexp/syntax.Compile
         0     0%   100%    13.44GB 31.99%  regexp/syntax.Parse (inline)
```

```bash
$ go tool pprof mem.prof
```
```bash
File: hw10_program_optimization.test
Build ID: 1abefdb2009a1e6b4e482c1b5829f1dc53029a9b
Type: alloc_space
Time: 2026-03-18 22:01:41 MSK
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) list countDomains
Total: 42.01GB
ROUTINE ======================== github.com/fixme_my_friend/hw10_program_optimization.countDomains in /mnt/u/projects/Go_Training/diasoft-go-course/hw10_program_optimization/stats.go
  240.50MB    36.42GB (flat, cum) 86.69% of Total
         .          .     50:func countDomains(u users, domain string) (DomainStat, error) {
         .          .     51:   result := make(DomainStat)
         .          .     52:
         .          .     53:   for _, user := range u {
  240.50MB    36.42GB     54:           matched, err := regexp.Match("\\."+domain, []byte(user.Email))
         .          .     55:           if err != nil {
         .          .     56:                   return nil, err
         .          .     57:           }
         .          .     58:
         .          .     59:           if matched {
(pprof)
```

Сохраняем начальные значения
```bash
go test -bench=. -benchmem | tee old.txt
```

-----------------------------------------------------------
##  Стало
```bash
$ go tool pprof -text -nodecount=20 
-alloc_space mem.prof
File: hw10_program_optimization.test
Build ID: fb0a28968b11d5d755434b62f8aea8d6ca7d2d0b
Type: alloc_space
Time: 2026-03-20 16:07:16 MSK
Showing nodes accounting for 6.94GB, 99.09% of 7.01GB total
Dropped 50 nodes (cum <= 0.04GB)
Showing top 20 nodes out of 32
      flat  flat%   sum%        cum   cum%
    3.90GB 55.60% 55.60%     3.90GB 55.60%  bufio.(*Scanner).Scan
    0.92GB 13.06% 68.66%     0.92GB 13.06%  bytes.NewBufferString (inline)
    0.56GB  8.02% 76.69%     0.56GB  8.02%  regexp/syntax.(*compiler).inst (inline)
    0.44GB  6.29% 82.98%     0.44GB  6.29%  github.com/mailru/easyjson/jlexer.(*Lexer).String
    0.31GB  4.43% 87.41%     0.31GB  4.43%  regexp/syntax.(*parser).newRegexp (inline)
    0.27GB  3.80% 91.20%     6.09GB 86.85%  github.com/fixme_my_friend/hw10_program_optimization.countDomains
    0.17GB  2.43% 93.63%     1.24GB 17.72%  regexp.compile
    0.10GB  1.44% 95.06%     0.45GB  6.45%  regexp/syntax.parse
    0.09GB  1.26% 96.33%     0.09GB  1.26%  regexp.(*bitState).reset
    0.09GB  1.22% 97.55%     0.09GB  1.22%  strings.genSplit
    0.05GB  0.67% 98.21%     0.05GB  0.67%  strings.(*Builder).grow
    0.04GB  0.57% 98.79%     0.09GB  1.22%  regexp/syntax.(*compiler).init (inline)
    0.02GB  0.31% 99.09%     0.04GB  0.59%  regexp/syntax.(*parser).push
         0     0% 99.09%     0.44GB  6.29%  github.com/fixme_my_friend/hw10_program_optimization.(*User).UnmarshalJSON (inline)  
         0     0% 99.09%        7GB 99.92%  github.com/fixme_my_friend/hw10_program_optimization.BenchmarkGetDomainStat
         0     0% 99.09%     6.09GB 86.85%  github.com/fixme_my_friend/hw10_program_optimization.GetDomainStat
         0     0% 99.09%     0.44GB  6.29%  github.com/fixme_my_friend/hw10_program_optimization.easyjsonE3ab7953DecodeGithubComFixmeMyFriendHw10ProgramOptimization
         0     0% 99.09%     0.11GB  1.56%  regexp.(*Regexp).MatchString (inline)
         0     0% 99.09%     0.11GB  1.56%  regexp.(*Regexp).backtrack
         0     0% 99.09%     0.11GB  1.56%  regexp.(*Regexp).doExecute
```


```bash
$ go tool pprof mem.prof
```
```bash
File: hw10_program_optimization.test
Build ID: fb0a28968b11d5d755434b62f8aea8d6ca7d2d0b
Type: alloc_space
Time: 2026-03-20 16:07:16 MSK
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) list countDomains
Total: 7.01GB
ROUTINE ======================== github.com/fixme_my_friend/hw10_program_optimization.countDomains in /mnt/u/projects/Go_Training/diasoft-go-course/hw10_program_optimization/stats.go
  272.54MB     6.09GB (flat, cum) 86.85% of Total
         .          .     30:func countDomains(scanner *bufio.Scanner, domain string) (DomainStat, error) {
   44.50MB    44.50MB     31:   result := make(DomainStat)
         .          .     32:   // 1 компиляция regexp
   17.50MB     1.26GB     33:   re, err := regexp.Compile(`\.` + domain)
         .          .     34:   if err != nil {
         .          .     35:           return nil, err
         .          .     36:   }
         .     3.90GB     37:   for scanner.Scan() {
         .          .     38:           line := scanner.Bytes()
         .          .     39:           if len(line) == 0 {
         .          .     40:                   continue
         .          .     41:           }
         .          .     42:
         .          .     43:           var user User
         .          .     44:           // easyjson вместо encoding/json
         .   451.51MB     45:           if err := user.UnmarshalJSON(line); err != nil {
         .          .     46:                   return nil, fmt.Errorf("unmarshal error: %w", err)
         .          .     47:           }
         .   112.04MB     48:           matched := re.MatchString(user.Email)
         .          .     49:           if matched {
         .   135.50MB     50:                   domainPart := strings.ToLower(strings.SplitN(user.Email, "@", 2)[1])
  210.54MB   210.54MB     51:                   result[domainPart]++ // Вместо двух вызовов
         .          .     52:           }
         .          .     53:   }
         .          .     54:   return result, nil
         .          .     55:}
(pprof) 
```

Сохраняем начальные значения
```bash
go test -bench=. -benchmem | tee new.txt
```

## Сравнение

```bash
$ benchstat old.txt new.txt
goos: linux
goarch: amd64
pkg: github.com/fixme_my_friend/hw10_program_optimization
cpu: AMD Ryzen 9 9950X 16-Core Processor
                 │     old.txt     │             new.txt             │
                 │     sec/op      │    sec/op     vs base           │
GetDomainStat-32   59228.71µ ± ∞ ¹   10.50µ ± ∞ ¹  ~ (p=1.000 n=1) ²
¹ need >= 6 samples for confidence interval at level 0.95
² need >= 4 samples to detect a difference at alpha level 0.05

                 │      old.txt       │             new.txt              │
                 │        B/op        │     B/op       vs base           │
GetDomainStat-32   153180.680Ki ± ∞ ¹   7.198Ki ± ∞ ¹  ~ (p=1.000 n=1) ²
¹ need >= 6 samples for confidence interval at level 0.95
² need >= 4 samples to detect a difference at alpha level 0.05

                 │     old.txt      │            new.txt             │
                 │    allocs/op     │  allocs/op   vs base           │
GetDomainStat-32   1700093.00 ± ∞ ¹   58.00 ± ∞ ¹  ~ (p=1.000 n=1) ²
¹ need >= 6 samples for confidence interval at level 0.95
² need >= 4 samples to detect a difference at alpha level 0.05
```