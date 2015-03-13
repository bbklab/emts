<?php
/* vim: set expandtab tabstop=4 shiftwidth=4 softtabstop=4 foldmethod=marker: */

/**
 * eYou Mail lib
 * 
 * @category   eYou_Mail
 * @package    Em_
 * @copyright  $_EYOUMBR_COPYRIGHT_$
 * @version    $_EYOUMBR_VERSION$_$
 */

/**
 * license_view 
 * 
 * @category   eYou_Mail
 * @package    Em_Core
 * @subpackage Em_Core
 */
class license_view
{
    // {{{ consts

    /**
     * 调试标识 
     *
     * @const bool
     */
    const DEBUG = false;

    /**
     * 邮件版本标识 
     * 
     * @const int
     */
    const MAIL_VER_1 = 10;
    const MAIL_VER_2 = 20;

    // }}}
    // {{{ functions 
    // {{{ public function show()

    /**
     * 显示 license 信息 
     * 
     * @return string
     */
    public function show()
    {
        if (defined('EMBASE_PATH_EYOU_MAIL_CONF')) {
            require_once EMBASE_PATH_EYOU_MAIL_CONF . 'conf_global.php';
        } else {
            require_once 'conf_global.php';
        }

        if (!file_exists(PATH_EYOUM_LIB . 'em_core.class.php')) {
            echo '';
            return;
        }

        $core_vl = em_core::vl();

        $license = array();
        switch ($this->_check_version()) {
        case self::MAIL_VER_1:
            $license = $this->_get_lincense_ver_1($core_vl);
            break;

        case self::MAIL_VER_2:
            $license = $this->_get_lincense_ver_2($core_vl);
            break;
        }

        if (!self::DEBUG) {
            echo json_encode($license);
        } else {
            print_r($license);
        }
    }

    // }}}
    // {{{ protected function _check_version()

    /**
     * _check_version 
     * 
     * @return void
     * @throws em_exception
     */
    protected function _check_version()
    {
        return function_exists('em_cl_m5') ? self::MAIL_VER_2 : self::MAIL_VER_1;
    }

    // }}}
    // {{{ protected function _get_lincense_ver_1()

    /**
     * 获取 v1 版本 license 信息 
     * 
     * @param int $vl 
     * @return array
     */
    protected function _get_lincense_ver_1($vl)
    {
        $license_type = array(
            '0' => gettext('formal business license'),
            '1' => gettext('demo business license'),
            '99' => gettext('trial version'),
            '-1' => gettext('invalid business license'),
            '-2' => gettext('expired business license'),
            '-3' => gettext('invalid business license'),
        );

        $license_base = array(
            'version'   => (defined('EYOUM_VERSION')) ? EYOUM_VERSION : '',
            'serial'    => $this->_get_serial(),
        );

        switch ($vl) {
        // 无效的
        case '-1':
        case '-3':
            return array_merge($license_base, array(
                'type'  => $license_type[$vl]
            ));
            break;

            // 正式
        case '0':
            $license_acct_num = self::_config('license_acct_num');
            $license_module = self::_config('license_module');
            $license_init_time = self::_config('license_init_time');
            $license_end_time = gettext('forever');
            break;

            // 演示
        case '1':
            // 过期的
        case '-2':
            $license_acct_num = self::_config('license_acct_num');
            $license_module = self::_config('license_module');
            $license_init_time = self::_config('license_init_time');
            $license_end_time = date('Y/m/d', strtotime('+3 month', strtotime($license_init_time)));
            break;

            // 试用版
        case '99':
            $license_acct_num = self::_config('license_acct_num');
            $license_module = self::_config('license_module');

            return array_merge($license_base, array(
                'type'      => $license_type[$vl],
                'user_num'  => $license_acct_num,
                'module'    => $this->_format_module($license_module)
            ));
            break;
        }

        return array_merge($license_base, array(
            'type'      => $license_type[$vl],
            'start_time' => $license_init_time,
            'end_time'  => $license_end_time,
            'user_num'  => $license_acct_num,
            'module'    => $this->_format_module($license_module)
        ));
    }

    // }}}
    // {{{ protected staic function _config()

    protected static function _config($config)
    {
        if (method_exists('em_config', 'get')) {
            return em_config::get($config);
        } elseif (method_exists('em_config', 'get_config')) {
            return em_config::get_config($config);
        } else {
            return '';
        }
    }

    // }}}
    // {{{ protected function _get_lincense_ver_2()

    /**
     * 获取 v2 版本 license 信息 
     * 
     * @param array $vl 
     * @return array
     */
    protected function _get_lincense_ver_2($vl)
    {
        $license_type = array(
            '-1' => gettext('invalid business license'),
            '0' => gettext('formal business license'),
            '1' => gettext('trial business license'),
            '2' => gettext('demo business license'),
        );

        $allocated_acct_num = em_core::get_license_allocated_acct_num();
        $allocated_alias_num = em_core::get_license_allocated_alias_num();
        $remain_acct_num = $vl['user_num'] - $allocated_acct_num;
        $remain_alias_num = $vl['alias_num'] - $allocated_alias_num;

        return array(
            'version' => $vl['version'],
            'domain'    => $vl['domain'],
            'type'      => $license_type[$vl['type']],
            'serial' => $this->_get_serial(),
            'user_num'  => number_format($vl['user_num']),
            'alias_num' => number_format($vl['alias_num']),
            'allocated_acct_num' => number_format($allocated_acct_num),
            'allocated_alias_num' => number_format($allocated_alias_num),
            'remain_acct_num' => ($remain_acct_num < 0) ? 0 : number_format($remain_acct_num),
            'remain_alias_num' => ($remain_alias_num < 0) ? 0 : number_format($remain_alias_num),
            'start_time' => date('Y/m/d', $vl['start_time']),
            'end_time'  => date('Y/m/d', $vl['end_time']),
            'is_over'   => $vl['is_over'],
            'module'    => $this->_format_module($vl['module']),
        );
    }

    // }}}
    // {{{ protected function _format_module()

    /**
     * 格式化增值模块
     *
     * @param Array $module 模块数组
     */
    protected function _format_module($module) {
        if (empty($module)) {
            return '';
        }

        $result = array();
        foreach ($module as $value) {
            if (is_array($value)) {
                foreach ($value as $key => $sub_value) {
                    $result[$key] = $sub_value;
                }
            } else {
                $result[$value] = gettext('module ' . $value);
            }
        }

        return $result;
    }

    // }}}
    // {{{ protected function _get_serial()

    /*
     * 生成机器码
     *
     * @return json
     */
    protected function _get_serial() 
    {
        if (!file_exists(PATH_EYOUM_SBIN . '/em_serial')) {
            return '';
        }

        try {
            $serial = exec(PATH_EYOUM_SBIN . '/em_serial');
        } catch (em_exception $e) {
            return '';
        }

        return $serial;
    }

    // }}}
    // }}}
}

$view = new license_view;
$view->show();

