package jd_cookie

import (
	"fmt"
	"net/url"

	"github.com/cdle/sillyGirl/core"
	"github.com/cdle/sillyGirl/develop/qinglong"
)

func initRongQi() {
	core.AddCommand("", []core.Function{
		{
			Rules: []string{"迁移"},
			Admin: true,
			Handle: func(s core.Sender) interface{} {
				if it := s.GetImType(); it != "terminal" && it != "tg" {
					return "可能会产生大量消息，请在终端或tg进行操作。"
				}
				//容器内去重
				var memvs = map[*qinglong.QingLong][]qinglong.Env{} //分组记录ck
				var aggregated = []*qinglong.QingLong{}
				for _, ql := range qinglong.GetQLS() {
					if ql.AggregatedMode {
						aggregated = append(aggregated, ql)
					}
					envs, err := qinglong.GetEnvs(ql, "JD_COOKIE")
					if err == nil {
						var mc = map[string]bool{}
						nn := []qinglong.Env{}
						for _, env := range envs {
							if env.Status == 0 {
								env.PtPin = core.FetchCookieValue(env.Value, "pt_pin")
								if env.PtPin == "" {
									continue
								}
								name, _ = url.QueryUnescape(env.PtPin)
								if _, ok := mc[env.PtPin]; ok {
									if _, err := qinglong.Req(ql, qinglong.PUT, qinglong.ENVS, "/disable", []byte(`["`+env.ID+`"]`)); err == nil {
										s.Reply(fmt.Sprintf("在同一容器发现到重复账号，已隐藏(%s)%s。", name, ql.GetTail()))
									}
									env.Remarks = "重复账号。"
									qinglong.UdpEnv(ql, env)
								} else {
									mc[env.PtPin] = true
									nn = append(nn, env)
								}
							}
						}
						memvs[ql] = nn
					}
				}
				//容器间去重
				var eql = map[string]*qinglong.QingLong{}
				for ql, envs := range memvs {
					if ql.AggregatedMode {
						continue
					}
					nn := []qinglong.Env{}
					for _, env := range envs {
						name, _ = url.QueryUnescape(env.PtPin)
						if _, ok := eql[env.PtPin]; ok {
							if ql_, err := qinglong.Req(ql, qinglong.PUT, qinglong.ENVS, "/disable", []byte(`["`+env.ID+`"]`)); err == nil {
								s.Reply(fmt.Sprintf("在%s发现重复账号，已隐藏(%s)%s。", ql.GetName(), name, ql_.GetTail()))
							}
							env.Remarks = "重复账号。"
							qinglong.UdpEnv(ql, env)
						} else {
							eql[env.PtPin] = ql
							nn = append(nn, env)
						}
					}
					memvs[ql] = nn
				}
				//聚合
				if len(aggregated) > 0 {
					for _, aql := range aggregated {
						toapp := []qinglong.Env{}
						for ql, envs := range memvs {
							toapp_ := []qinglong.Env{}
							if ql == aql {
								continue
							}
							for _, env := range envs {
								if !envContain(append(memvs[aql], toapp...), env) {
									toapp = append(toapp, env)
									toapp_ = append(toapp_, env)
								}
							}
							if len(toapp_) > 0 {
								memvs[aql] = append(memvs[aql], toapp_...)
								if err := qinglong.AddEnv(aql, toapp_...); err != nil {
									s.Reply(fmt.Sprintf("失败转移%d个账号到聚合容器%s：%v%s", len(toapp_), aql.GetName(), err, ql.GetName()))
								} else {
									s.Reply(fmt.Sprintf("成功转移%d个账号到聚合容器%s。%s", len(toapp_), aql.GetName(), ql.GetName()))
								}
							}
						}

					}
				}
				return "迁移完成。"
			},
		},
	})
}

func envContain(ay []qinglong.Env, e qinglong.Env) bool {
	for _, v := range ay {
		if v.PtPin == e.PtPin {
			return true
		}
	}
	return false
}
