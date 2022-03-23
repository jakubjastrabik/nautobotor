package nautobotor

import (
	"github.com/coredns/coredns/plugin/test"

	"github.com/miekg/dns"
)

// these examples don't have an additional opt RR set, because that's gets added by the server.
var wildcardTestCases = []test.Case{
	{
		Qname: "wild.dnssex.nl.", Qtype: dns.TypeTXT,
		Answer: []dns.RR{
			test.TXT(`wild.dnssex.nl.	1800	IN	TXT	"Doing It Safe Is Better"`),
		},
		Ns: dnssexAuth[:len(dnssexAuth)-1], // remove RRSIG on the end
	},
	{
		Qname: "a.wild.dnssex.nl.", Qtype: dns.TypeTXT,
		Answer: []dns.RR{
			test.TXT(`a.wild.dnssex.nl.	1800	IN	TXT	"Doing It Safe Is Better"`),
		},
		Ns: dnssexAuth[:len(dnssexAuth)-1], // remove RRSIG on the end
	},
	{
		Qname: "wild.dnssex.nl.", Qtype: dns.TypeTXT, Do: true,
		Answer: []dns.RR{
			test.RRSIG("wild.dnssex.nl.	1800	IN	RRSIG	TXT 8 2 1800 20160428190224 20160329190224 14460 dnssex.nl. FUZSTyvZfeuuOpCm"),
			test.TXT(`wild.dnssex.nl.	1800	IN	TXT	"Doing It Safe Is Better"`),
		},
		Ns: append([]dns.RR{
			test.NSEC("a.dnssex.nl.	14400	IN	NSEC	www.dnssex.nl. A AAAA RRSIG NSEC"),
			test.RRSIG("a.dnssex.nl.	14400	IN	RRSIG	NSEC 8 3 14400 20160428190224 20160329190224 14460 dnssex.nl. S+UMs2ySgRaaRY"),
		}, dnssexAuth...),
	},
	{
		Qname: "a.wild.dnssex.nl.", Qtype: dns.TypeTXT, Do: true,
		Answer: []dns.RR{
			test.RRSIG("a.wild.dnssex.nl.	1800	IN	RRSIG	TXT 8 2 1800 20160428190224 20160329190224 14460 dnssex.nl. FUZSTyvZfeuuOpCm"),
			test.TXT(`a.wild.dnssex.nl.	1800	IN	TXT	"Doing It Safe Is Better"`),
		},
		Ns: append([]dns.RR{
			test.NSEC("a.dnssex.nl.	14400	IN	NSEC	www.dnssex.nl. A AAAA RRSIG NSEC"),
			test.RRSIG("a.dnssex.nl.	14400	IN	RRSIG	NSEC 8 3 14400 20160428190224 20160329190224 14460 dnssex.nl. S+UMs2ySgRaaRY"),
		}, dnssexAuth...),
	},
	// nodata responses
	{
		Qname: "wild.dnssex.nl.", Qtype: dns.TypeSRV,
		Ns: []dns.RR{
			test.SOA(`dnssex.nl.	1800	IN	SOA	linode.atoom.net. miek.miek.nl. 1459281744 14400 3600 604800 14400`),
		},
	},
	{
		Qname: "wild.dnssex.nl.", Qtype: dns.TypeSRV, Do: true,
		Ns: []dns.RR{
			// TODO(miek): needs closest encloser proof as well? This is the wrong answer
			test.NSEC(`*.dnssex.nl.	14400	IN	NSEC	a.dnssex.nl. TXT RRSIG NSEC`),
			test.RRSIG(`*.dnssex.nl.	14400	IN	RRSIG	NSEC 8 2 14400 20160428190224 20160329190224 14460 dnssex.nl. os6INm6q2eXknD5z8TaaDOV+Ge/Ko+2dXnKP+J1fqJzafXJVH1F0nDrcXmMlR6jlBHA=`),
			test.RRSIG(`dnssex.nl.	1800	IN	RRSIG	SOA 8 2 1800 20160428190224 20160329190224 14460 dnssex.nl. CA/Y3m9hCOiKC/8ieSOv8SeP964Bq++lyH8BZJcTaabAsERs4xj5PRtcxicwQXZiF8fYUCpROlUS0YR8Cdw=`),
			test.SOA(`dnssex.nl.	1800	IN	SOA	linode.atoom.net. miek.miek.nl. 1459281744 14400 3600 604800 14400`),
		},
	},
}

var dnssexAuth = []dns.RR{
	test.NS("dnssex.nl.	1800	IN	NS	linode.atoom.net."),
	test.NS("dnssex.nl.	1800	IN	NS	ns-ext.nlnetlabs.nl."),
	test.NS("dnssex.nl.	1800	IN	NS	omval.tednet.nl."),
	test.RRSIG("dnssex.nl.	1800	IN	RRSIG	NS 8 2 1800 20160428190224 20160329190224 14460 dnssex.nl. dLIeEvP86jj5ndkcLzhgvWixTABjWAGRTGQsPsVDFXsGMf9TGGC9FEomgkCVeNC0="),
}
