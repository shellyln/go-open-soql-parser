package parser_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/shellyln/go-open-soql-parser/soql/parser"
	"github.com/shellyln/go-open-soql-parser/soql/parser/types"
)

func TestParse(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name     string
		args     args
		want     interface{}
		wantErr  bool
		dbgBreak bool
	}{{
		name:    "1",
		args:    args{s: `SELECT CONCAT(1,2,3,4) FROM Contact`},
		want:    nil,
		wantErr: false,
	}, {
		name:    "left join 1",
		args:    args{s: `SELECT Id FROM Contact WHERE LastName = 'foo' or Account.Name = 'bar'`},
		want:    nil,
		wantErr: false,
	}, {
		name:    "left join 2",
		args:    args{s: `SELECT Id, Account.Id, Account.Name FROM Contact ORDER BY Account.Name`},
		want:    nil,
		wantErr: false,
	}, {
		name:    "inner join 1",
		args:    args{s: `SELECT Id FROM Contact WHERE LastName = 'foo' and Account.Name = 'bar'`},
		want:    nil,
		wantErr: false,
	}, {
		name:    "inner join 2",
		args:    args{s: `SELECT Id FROM Contact WHERE Account.Name = 'bar'`},
		want:    nil,
		wantErr: false,
	}, {
		name:    "fieldset 1",
		args:    args{s: `SELECT FIELDS(all) FROM Contact`},
		want:    nil,
		wantErr: false,
	}, {
		name: "subquery 1",
		args: args{s: `SELECT acc.Id,(SELECT Id,acc.Id FROM con.Departments) qwerty FROM Contact con`}, // pass (ok) <- each acc.id are independent
		// args:    args{s: `SELECT acc.Id,(SELECT Id,acc.Id FROM con.Departments) qwerty FROM Contact con, Account acc`}, // <- error (ok) (The siblings object item is not allowed to be selected)
		// args:    args{s: `SELECT acc.Id,(SELECT Id,contact.acc.Id FROM Departments) qwerty FROM Contact con`}, // <- error (ok) (The siblings object item is not allowed to be selected)
		// args:    args{s: `SELECT acc.Id,(SELECT Id,con.acc.Id FROM Departments) qwerty FROM Contact con`}, // <- error (ok) (The siblings object item is not allowed to be selected)
		// args:    args{s: `SELECT id qwerty FROM Contact con where x in (SELECT Id FROM Departments where contact=con.id)`}, // <- pass (ng?) // BUG:
		// args:    args{s: `SELECT id qwerty FROM Contact con where x in (SELECT Id FROM Departments where contact=contact.id)`}, // <- pass (ng?) // BUG:
		// args:    args{s: `SELECT acc.Id,(SELECT Id,account.Id FROM con.acc.Departments) qwerty FROM Contact con, Account acc`}, // BUG: ([FATAL]Internal error: Source view is not found: Contact.Account) // makeJoinSubqueriesTasks leftObject // key was removed when joint
		want:    nil,
		wantErr: false,
	}, {
		name:    "co-related subquery 1",
		args:    args{s: `SELECT (SELECT Id FROM con.Departments where contact=contact.id) qwerty FROM Contact con`},
		want:    nil,
		wantErr: false,
	}, {
		name:    "co-related subquery 2",
		args:    args{s: `SELECT (SELECT Id FROM con.Departments where contact=con.id) qwerty FROM Contact con`},
		want:    nil,
		wantErr: false,
	}, {
		name:    "co-related subquery 3 (non co-related)",
		args:    args{s: `SELECT (SELECT Id, con.Id FROM con.Departments) qwerty FROM Contact con`},
		want:    nil,
		wantErr: true,
	}, {
		name:    "fieldset 1",
		args:    args{s: `SELECT fields(acc.all) FROM Contact con, con.Account acc`},
		want:    nil,
		wantErr: false,
	}, {
		name:    "fieldset 2",
		args:    args{s: `SELECT fields(con.acc.all) FROM Contact con, con.Account acc`},
		want:    nil,
		wantErr: true,
	}, {
		name:     "cond subquery",
		args:     args{s: `SELECT Id FROM Contact WHERE Name in (select asd.x from qwe)`},
		want:     nil,
		wantErr:  false,
		dbgBreak: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.dbgBreak {
				t.Log("debug")
			}

			got, err := parser.Parse(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				// t.Errorf("Parse() = %v, want %v", got, tt.want)
				// return
				t.Logf("Parse() = %v, want %v", got, tt.want)
			}

			jsonBytes1, err := json.Marshal(got)
			if err != nil {
				t.Errorf("json.Marshal() (1) error = %v", err)
				return
			}
			var unmarshal1 types.SoqlQuery
			err = json.Unmarshal(jsonBytes1, &unmarshal1)
			if err != nil {
				t.Errorf("json.Unmarshal() (1) error = %v", err)
				return
			}

			jsonBytes2, err := json.Marshal(unmarshal1)
			if err != nil {
				t.Errorf("json.Marshal() (2) error = %v", err)
				return
			}
			var unmarshal2 types.SoqlQuery
			err = json.Unmarshal(jsonBytes2, &unmarshal2)
			if err != nil {
				t.Errorf("json.Unmarshal() (2) error = %v", err)
				return
			}
			if string(jsonBytes1) != string(jsonBytes2) {
				t.Errorf("Marshal(1) = %v, Marshal(2) %v", string(jsonBytes1), string(jsonBytes2))
				return
			}

			jsonBytes3, err := json.Marshal(unmarshal2)
			if err != nil {
				t.Errorf("json.Marshal() (3) error = %v", err)
				return
			}
			if string(jsonBytes1) != string(jsonBytes3) {
				t.Errorf("Marshal(1) = %v, Marshal(3) %v", string(jsonBytes1), string(jsonBytes3))
				return
			}
		})
	}
}

func TestParse2(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name     string
		args     args
		want     interface{}
		wantErr  bool
		dbgBreak bool
	}{{
		name: "1",
		args: args{s: `
			SELECT
			    acc.Id xid
			  --, con.foo__r.xxx
			  , foo__r.bar__r.zzz
			  , foo__r.yyy
			  , con.Name xname
			  , con.acc.ddd xddd
			  , CONCAT(TRIM(acc.Name), '/', TRIM(con.Name), 123.45, 0xacc0) cname
			  , FLAT(acc.Name)
			  , (SELECT Id FROM con.Departments where uuu=con.Zzz and vvv=con.Id) qwerty
			  , (select Id from r3.lkjh where name='www')
			FROM
			    Contact con
			  , con.Account acc
			  , PPP.QQQ.RRR r3
			WHERE
			    not (Name like 'a%' or Name like 'b%')
				and
				acc.Name in ('a', 'b', 'c', null)
				and
				acc.Id in ('a', 'b', 'c', null)
				and
				r3.Name in (select x,Id,Name,(select w from ghjksfd) from Contact)
				and
				Name > 0001-01-02
				and
				(((Name > 0001-01-02T01:01:01.123456789Z)
				or
				Name = :param1))
				and
				con.Name = acc.Name
				and
				LEN(con.Name) > 0
				and
				foo__r.bar__r.zzz = 1
			ORDER BY
			    acc.Name desc nulls last
			  --, acc.Id desc nulls last
			  , xid
			  , con.Name
			OFFSET 1000 LIMIT 100
			FOR update viewstat, tracking
		`},
		want:    nil,
		wantErr: false,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.dbgBreak {
				t.Log("debug")
			}

			got, err := parser.Parse(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			{
				s := got.Meta.ElapsedTime.String()
				t.Log(s)
			}

			if !reflect.DeepEqual(got, tt.want) {
				// t.Errorf("Parse() = %v, want %v", got, tt.want)
				// return
				t.Logf("Parse() = %v, want %v", got, tt.want)
			}

			jsonBytes1, err := json.Marshal(got)
			if err != nil {
				t.Errorf("json.Marshal() (1) error = %v", err)
				return
			}
			var unmarshal1 types.SoqlQuery
			err = json.Unmarshal(jsonBytes1, &unmarshal1)
			if err != nil {
				t.Errorf("json.Unmarshal() (1) error = %v", err)
				return
			}

			jsonBytes2, err := json.Marshal(unmarshal1)
			if err != nil {
				t.Errorf("json.Marshal() (2) error = %v", err)
				return
			}
			var unmarshal2 types.SoqlQuery
			err = json.Unmarshal(jsonBytes2, &unmarshal2)
			if err != nil {
				t.Errorf("json.Unmarshal() (2) error = %v", err)
				return
			}
			if string(jsonBytes1) != string(jsonBytes2) {
				t.Errorf("Marshal(1) = %v, Marshal(2) %v", string(jsonBytes1), string(jsonBytes2))
				return
			}

			jsonBytes3, err := json.Marshal(unmarshal2)
			if err != nil {
				t.Errorf("json.Marshal() (3) error = %v", err)
				return
			}
			if string(jsonBytes1) != string(jsonBytes3) {
				t.Errorf("Marshal(1) = %v, Marshal(3) %v", string(jsonBytes1), string(jsonBytes3))
				return
			}
		})
	}
}

func TestParse3(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name     string
		args     args
		want     interface{}
		wantErr  bool
		dbgBreak bool
	}{{
		name: "1",
		args: args{s: `
			SELECT
			    acc.Id xid
			  --, con.foo__r.xxx
			  , foo__r.bar__r.zzz
			  , foo__r.yyy
			  , con.Name xname
			  , con.acc.ddd xddd
			  , CONCAT(TRIM(acc.Name), '/', TRIM(con.Name), 123.45, 0xacc0) cname
			  , FLAT(acc.Name)
			FROM
			    Contact con
			  , con.Account acc
			  , PPP.QQQ.RRR r3
			WHERE
			    not (Name like 'a%' or Name like 'b%')
				and
				acc.Name in ('a', 'b', 'c', null)
				and
				acc.Id in ('a', 'b', 'c', null)
				and
				r3.Name in (select x,Id,Name from Contact)
				and
				Name > 0001-01-02
				and
				(((Name > 0001-01-02T01:01:01.123456789Z)
				or
				Name = :param1))
				and
				con.Name = acc.Name
				and
				LEN(con.Name) > 0
			GROUP BY
			    acc.Name
			  --, acc.Id
			  , xid
			  , con.Name
			  , foo__r.bar__r.zzz
			  , foo__r.yyy
			  , con.acc.ddd
			HAVING
			    LEN(MAX(con.Name)) > FOO(0)
				and
			    LEN(MAX(con.Id)) > 0
			ORDER BY
			    acc.Name desc nulls last
			  --, acc.Id desc nulls last
			  , xid
			  , con.Name
			OFFSET 1000 LIMIT 100
			FOR update viewstat, tracking
		`},
		want:    nil,
		wantErr: false,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.dbgBreak {
				t.Log("debug")
			}

			got, err := parser.Parse(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			{
				s := got.Meta.ElapsedTime.String()
				t.Log(s)
			}

			if !reflect.DeepEqual(got, tt.want) {
				// t.Errorf("Parse() = %v, want %v", got, tt.want)
				// return
				t.Logf("Parse() = %v, want %v", got, tt.want)
			}

			jsonBytes1, err := json.Marshal(got)
			if err != nil {
				t.Errorf("json.Marshal() (1) error = %v", err)
				return
			}
			var unmarshal1 types.SoqlQuery
			err = json.Unmarshal(jsonBytes1, &unmarshal1)
			if err != nil {
				t.Errorf("json.Unmarshal() (1) error = %v", err)
				return
			}

			jsonBytes2, err := json.Marshal(unmarshal1)
			if err != nil {
				t.Errorf("json.Marshal() (2) error = %v", err)
				return
			}
			var unmarshal2 types.SoqlQuery
			err = json.Unmarshal(jsonBytes2, &unmarshal2)
			if err != nil {
				t.Errorf("json.Unmarshal() (2) error = %v", err)
				return
			}
			if string(jsonBytes1) != string(jsonBytes2) {
				t.Errorf("Marshal(1) = %v, Marshal(2) %v", string(jsonBytes1), string(jsonBytes2))
				return
			}

			jsonBytes3, err := json.Marshal(unmarshal2)
			if err != nil {
				t.Errorf("json.Marshal() (3) error = %v", err)
				return
			}
			if string(jsonBytes1) != string(jsonBytes3) {
				t.Errorf("Marshal(1) = %v, Marshal(3) %v", string(jsonBytes1), string(jsonBytes3))
				return
			}
		})
	}
}
