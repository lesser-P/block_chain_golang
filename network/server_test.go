package network

import "testing"

func TestBytes(t *testing.T) {
	t.Log("测试命令拼接、拆分功能")
	{
		t.Log("\t测试拼接功能：")
		v := version{versionInfo, 10, ""}
		b := jointMessage(cVersion, v.serialize())
		t.Log("\t拼接后的字节数组为：", b)
		t.Log("\t测试拆分功能：")
		cmd, content := splitMessage(b)
		newV := version{}
		newV.deserialize(content)
		t.Log("\t命令为：", cmd, "长度：", len(cmd))
		t.Log("\t版本信息：", newV)
	}
}
