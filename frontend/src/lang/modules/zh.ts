export default {
    commons: {
        button: {
            create: '新建',
            create: '创建',
            add: '添加',
            delete: '删除',
            edit: '编辑',
            confirm: '确认',
            cancel: '取消',
            reset: '重置',
            login: '登陆',
            conn: '连接',
        },
        search: {
            timeStart: '开始时间',
            timeEnd: '结束时间',
            timeRange: '至',
            dateStart: '开始日期',
            dateEnd: '结束日期',
        },
        table: {
            name: '名称',
            group: '组',
            createdAt: '创建时间',
            date: '时间',
            updatedAt: '更新时间',
            operate: '操作',
            message: '信息',
            description: '描述信息',
        },
        msg: {
            delete: '此操作不可回滚,是否继续',
            deleteTitle: '删除',
            deleteSuccess: '删除成功',
            loginSuccess: '登陆成功',
            operationSuccess: '操作成功',
            requestTimeout: '请求超时,请稍后重试',
            infoTitle: '提示',
            sureLogOut: '您是否确认退出登录?',
            createSuccess: '新建成功',
            updateSuccess: '更新成功',
        },
        login: {
            captchaHelper: '请输入验证码',
        },
        rule: {
            username: '请输入用户名',
            password: '请输入密码',
            requiredInput: '请填写必填项',
            requiredSelect: '请选择必选项',
            commonName: '支持英文、中文、数字、.-_,长度1-30',
            email: '邮箱格式错误',
            ip: '请输入正确的 IP 地址',
            port: '请输入正确的端口',
        },
        res: {
            paramError: '请求失败,请稍后重试!',
            forbidden: '当前用户无权限',
            serverError: '服务异常',
            notFound: '资源不存在',
            commonError: '请求失败',
        },
        header: {
            language: '国际化',
            zh: '简体中文',
            en: 'English',
            theme: '布局设置',
            globalTheme: '全局主题',
            themeColor: '主题颜色',
            darkTheme: '暗黑主题',
            personalData: '个人资料',
            changePassword: '修改密码',
            logout: '退出登录',
        },
    },
    business: {
        user: {
            username: '用户名',
            email: '邮箱',
            password: '密码',
        },
    },
    menu: {
        home: '概览',
        demo: '样例',
        terminal: '终端管理',
        monitor: '监控',
        operations: '操作记录',
        files: '文件管理',
    },
    home: {
        welcome: '欢迎使用',
    },
    tabs: {
        more: '更多',
        closeCurrent: '关闭当前',
        closeOther: '关闭其它',
        closeAll: '关闭所有',
    },
    header: {
        componentSize: '组件大小',
        language: '国际化',
        theme: '全局主题',
        layoutConfig: '布局设置',
        primary: 'primary',
        darkMode: '暗黑模式',
        greyMode: '灰色模式',
        weakMode: '色弱模式',
        fullScreen: '全屏',
        exitFullScreen: '退出全屏',
        personalData: '个人资料',
        changePassword: '修改密码',
        logout: '退出登录',
    },
    monitor: {
        avgLoad: '平均负载',
        loadDetail: '负载详情',
        resourceUsage: '资源使用率',
        min: '分钟',
        read: '读取',
        write: '写入',
        count: '次',
        readWriteCount: '读写次数',
        readWriteTime: '读写延迟',
        today: '今天',
        yestoday: '昨天',
        lastNDay: '近 {0} 天',
        memory: '内存',
        disk: '磁盘',
        network: '网络',
        up: '上行',
        down: '下行',
    },
    terminal: {
        conn: '连接',
        testConn: '连接测试',
        saveAndConn: '保存并连接',
        connTestOk: '连接信息可用',
        hostList: '主机信息',
        createConn: '创建连接',
        createGroup: '创建分组',
        expand: '全部展开',
        fold: '全部收缩',
        batchInput: '批量输入',
        quickCommand: '快速命令',
        groupDeleteHelper: '移除组后，组内所有连接将迁移到 default 组内，是否确认',
        quickCmd: '快捷命令',
        command: '命令',
        addHost: '添加主机',
        localhost: '本地服务器',
        name: '名称',
        port: '端口',
        user: '用户',
        authMode: '认证方式',
        passwordMode: '密码输入',
        keyMode: '密钥输入',
        password: '密码',
        key: '密钥',
        emptyTerminal: '暂无终端连接',
    },
    operations: {
        detail: {
            users: '用户',
            hosts: '主机',
            groups: '组',
            commands: '快捷命令',
            auth: '用户',
            post: '创建',
            put: '更新',
            update: '更新',
            delete: '删除',
            login: '登录',
            logout: '退出',
            del: '删除',
        },
        operatoin: '操作',
        status: '状态',
        request: '请求',
        response: '响应',
    },
    file: {
        dir: '文件夹',
        upload: '上传',
        download: '下载',
        fileName: '文件名',
        search: '查找',
        mode: '权限',
        owner: '所有者',
        file: '文件',
        remoteFile: '远程下载',
        share: '分享',
        sync: '数据同步',
        size: '大小',
        updateTime: '修改时间',
        open: '打开',
        rename: '重命名',
        role: '权限',
        info: '属性',
        linkFile: '软连接文件',
        terminal: '终端',
        shareList: '分享列表',
        zip: '压缩',
    },
};
