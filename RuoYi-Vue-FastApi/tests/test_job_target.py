"""
spec-09 补充：invoke_target 解析与校验测试（对齐java JobInvokeUtil）
"""
import sys
import os

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

import pytest
from module_task.target_resolver import parse_target, validate_target, _parse_param
from module_task.registry import register, is_registered


class TestParseParam:
    """参数字面量转python值（对齐java类型推断规则）"""

    def test_string_single_quote(self):
        assert _parse_param("'ry'") == 'ry'

    def test_string_double_quote(self):
        assert _parse_param('"hello"') == 'hello'

    def test_boolean(self):
        assert _parse_param('true') is True
        assert _parse_param('false') is False
        assert _parse_param('True') is True  # java equalsIgnoreCase

    def test_long_suffix(self):
        assert _parse_param('2000L') == 2000
        assert isinstance(_parse_param('2000L'), int)

    def test_double_suffix(self):
        assert _parse_param('316.50D') == 316.50
        assert isinstance(_parse_param('316.50D'), float)

    def test_int(self):
        assert _parse_param('100') == 100

    def test_invalid(self):
        with pytest.raises(ValueError):
            _parse_param('abc')


class TestParseTarget:
    """目标字符串解析（java预置3条任务全部覆盖）"""

    def test_no_params(self):
        bean, method, params = parse_target('ryTask.ryNoParams')
        assert (bean, method, params) == ('ryTask', 'ryNoParams', [])

    def test_single_string_param(self):
        bean, method, params = parse_target("ryTask.ryParams('ry')")
        assert (bean, method) == ('ryTask', 'ryParams')
        assert params == ['ry']

    def test_multiple_params(self):
        # java预置第3条任务的完整目标
        bean, method, params = parse_target(
            "ryTask.ryMultipleParams('ry', true, 2000L, 316.50D, 100)")
        assert (bean, method) == ('ryTask', 'ryMultipleParams')
        assert params == ['ry', True, 2000, 316.50, 100]

    def test_invalid_syntax(self):
        with pytest.raises(ValueError):
            parse_target('noDot')
        with pytest.raises(ValueError):
            parse_target('')


class TestValidateTarget:
    """校验链（对齐java SysJobController：黑名单+白名单）"""

    def _checker(self, name):
        return is_registered(name)

    def test_registered_pass(self):
        # ry_task 已注册（import触发）
        import module_task.ry_task  # noqa: F401
        assert validate_target('ryTask.ryNoParams', self._checker) == ''

    def test_rmi_blocked(self):
        assert 'rmi' in validate_target('task.rmi:x', self._checker)

    def test_ldap_blocked(self):
        assert 'ldap' in validate_target("task.ldap://evil", self._checker)

    def test_http_blocked(self):
        assert 'http' in validate_target('task.http://evil', self._checker)

    def test_forbidden_package(self):
        assert '违规' in validate_target('org.springframework.xx.doIt', self._checker)

    def test_not_in_whitelist(self):
        assert '白名单' in validate_target('unknownTask.doIt', self._checker)

    def test_unregistered_ry_method(self):
        import module_task.ry_task  # noqa: F401
        assert '白名单' in validate_target('ryTask.notExists', self._checker)


class TestRyTaskRegistered:
    """java预置的3条任务目标全部可解析且已注册"""

    def test_all_preset_jobs_registered(self):
        import module_task.ry_task  # noqa: F401
        from module_task.registry import get_task
        from module_task.target_resolver import parse_target

        preset = [
            'ryTask.ryNoParams',
            "ryTask.ryParams('ry')",
            "ryTask.ryMultipleParams('ry', true, 2000L, 316.50D, 100)",
        ]
        for target in preset:
            bean, method, params = parse_target(target)
            func = get_task(f'{bean}.{method}')
            assert callable(func), target
            assert validate_target(target, lambda n: is_registered(n)) == '', target


if __name__ == '__main__':
    sys.exit(pytest.main([__file__, '-v']))
