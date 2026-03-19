
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