export default {
    commons: {
        button: {
            create: 'Create',
            add: 'Add',
            save: 'Save',
            set: 'Reset',
            sync: 'Sync',
            delete: 'Delete',
            edit: 'Edit',
            enable: 'Enable',
            disable: 'Disable',
            confirm: 'Confirm',
            cancel: 'Cancel',
            reset: 'Reset',
            conn: 'Connect',
            clean: 'Clean',
            login: 'Login',
            close: 'Close',
            view: 'View',
            handle: 'Handle',
            expand: 'Expand',
            log: 'Log',
            saveAndEnable: 'Save and enable',
        },
        search: {
            timeStart: 'Time start',
            timeEnd: 'Time end',
            timeRange: 'To',
            dateStart: 'Date start',
            dateEnd: 'Date end',
        },
        table: {
            total: 'Total {0}',
            name: 'Name',
            type: 'Type',
            status: 'Status',
            statusSuccess: 'Success',
            statusFailed: 'Failed',
            records: 'Records',
            group: 'Group',
            createdAt: 'Creation Time',
            date: 'Date',
            updatedAt: 'Update Time',
            operate: 'Operations',
            message: 'Message',
            description: 'Description',
            interval: 'Interval',
        },
        msg: {
            delete: 'This operation cannot be rolled back. Do you want to continue',
            deleteTitle: 'Delete',
            deleteSuccess: 'Delete Success',
            loginSuccess: 'Login Success',
            operationSuccess: 'Successful operation',
            notSupportOperation: 'This operation is not supported',
            requestTimeout: 'The request timed out, please try again later',
            infoTitle: 'Hint',
            notRecords: 'No execution record is generated for the current task',
            sureLogOut: 'Are you sure you want to log out?',
            createSuccess: 'Create Success',
            updateSuccess: 'Update Success',
            uploadSuccess: 'Update Success',
        },
        login: {
            captchaHelper: 'Please enter the verification code',
            safeEntrance: 'Please use the correct entry to log in to the panel',
            reason: 'Cause of error:',
            reasonHelper:
                'At present, the newly installed machine has enabled the security entrance login. The newly installed machine will have a random 8-character security entrance name, which can also be modified in the panel Settings. If you do not record or do not remember, you can use the following methods to solve the problem',
            solution: 'The solution:',
            solutionHelper:
                'Run the following command on the SSH terminal to solve the problem: 1. View the /etc/init.d/bt default command on the panel',
            warnning:
                'Note: [Closing the security entrance] will make your panel login address directly exposed to the Internet, very dangerous, please exercise caution',
            codeInput: 'Please enter the 6-digit verification code of the MFA validator',
        },
        rule: {
            username: 'Please enter a username',
            password: 'Please enter a password',
            rePassword: 'The passwords are inconsistent. Please check and re-enter the password',
            requiredInput: 'Please enter the required fields',
            requiredSelect: 'Please select the required fields',
            commonName: 'Support English, Chinese, numbers, .-_, length 1-30',
            complexityPassword:
                'Please enter a password with more than 8 characters and must contain letters, digits, and special symbols',
            commonPassword: 'Please enter a password with more than 6 characters',
            email: 'Email format error',
            number: 'Please enter the correct number',
            ip: 'Please enter the correct IP address',
            port: 'Please enter the correct port',
        },
        res: {
            paramError: 'The request failed, please try again later!',
            forbidden: 'The current user has no permission',
            serverError: 'Service exception',
            notFound: 'The resource does not exist',
            commonError: 'The request failed',
        },
        header: {
            language: 'Internationalization',
            zh: '简体中文',
            en: 'English',
            theme: 'Layout Settings',
            globalTheme: 'Global',
            themeColor: 'Theme Color',
            darkTheme: 'Dark Theme',
            personalData: 'Personal Data',
            changePassword: 'Change Password',
            logout: 'Logout',
        },
    },
    auth: {
        username: 'Username',
        email: 'Email',
        password: 'Password',
    },
    menu: {
        home: 'Overview',
        apps: 'App Store',
        website: 'Website',
        project: 'Project',
        config: 'Config',
        firewall: 'Firewall',
        database: 'Database',
        container: 'Container',
        cronjob: 'Cronjob',
        host: 'Host',
        security: 'Security',
        files: 'File Management',
        monitor: 'Monitor',
        terminal: 'Terminal',
        settings: 'Setting',
        toolbox: 'Toolbox',
        operations: 'Operation Records',
    },
    home: {
        welcome: 'Welcome',
    },
    tabs: {
        more: 'More',
        closeCurrent: 'Close current',
        closeOther: 'Close other',
        closeAll: 'Close All',
    },
    header: {
        logout: 'Logout',
    },
    cronjob: {
        cronTask: 'Task',
        taskType: 'Task type',
        shell: 'shell',
        website: 'website',
        failedFilter: 'Failed Task Filtering',
        all: 'all',
        database: 'database',
        missBackupAccount: 'The backup account could not be found',
        syncDate: 'Synchronization time ',
        releaseMemory: 'Free memory',
        curl: 'Crul',
        taskName: 'Task name',
        cronSpec: 'Lifecycle',
        directory: 'Backup directory',
        sourceDir: 'Backup directory',
        exclusionRules: 'Exclusive rule',
        url: 'URL Address',
        target: 'Target',
        retainDays: 'Retain days',
        cronSpecRule: 'Please enter a correct lifecycle',
        perMonth: 'Per monthly',
        perWeek: 'Per week',
        perHour: 'Per hour',
        perNDay: 'Every N days',
        perNHour: 'Every N hours',
        perNMinute: 'Every N minutes',
        per: 'Every ',
        handle: 'Handle',
        day: 'Day',
        day1: 'Day',
        hour: ' Hour',
        minute: ' Minute',
        monday: 'Monday',
        tuesday: 'Tuesday',
        wednesday: 'Wednesday',
        thursday: 'Thursday',
        friday: 'Friday',
        saturday: 'Saturday',
        sunday: 'Sunday',
        shellContent: 'Script content',
        errRecord: 'Incorrect logging',
        errHandle: 'Task execution failure',
        noRecord: 'The execution did not generate any logs',
    },
    monitor: {
        avgLoad: 'Average load',
        loadDetail: 'Load detail',
        resourceUsage: 'Resource utilization rate',
        min: 'Minutes',
        read: 'Read',
        write: 'Write',
        count: 'Times',
        readWriteCount: 'Read or write Times',
        readWriteTime: 'Read or write delay',
        today: 'Today',
        yestoday: 'Yestoday',
        lastNDay: 'Last {0} day',
        memory: 'Memory',
        disk: 'Disk',
        network: 'Network',
        up: 'Up',
        down: 'Down',
    },
    terminal: {
        conn: 'connection',
        testConn: 'Test connection',
        saveAndConn: 'Save and Connect',
        connTestOk: 'Connection information available',
        hostList: 'Host information',
        createConn: 'Create a connection',
        createGroup: 'Create a group',
        expand: 'Expand all',
        fold: 'All contract',
        batchInput: 'Batch input',
        quickCommand: 'quick command',
        groupDeleteHelper:
            'After the group is removed, all connections in the group will be migrated to the default group. Confirm the information',
        quickCmd: 'Quick command',
        addHost: 'Add Host',
        localhost: 'Localhost',
        name: 'Name',
        port: 'Port',
        user: 'User',
        authMode: 'Auth Mode',
        passwordMode: 'password',
        keyMode: 'PrivateKey',
        password: 'Password',
        key: 'Private Key',
        emptyTerminal: 'No terminal is currently connected',
    },
    operations: {
        detail: {
            users: 'User',
            hosts: 'Host',
            groups: 'Group',
            commands: 'Command',
            backups: 'Backup Account',
            settings: 'Panel Setting',
            auth: 'User',
            login: ' login',
            logout: ' logout',
            post: ' create',
            put: ' update',
            update: ' update',
            delete: ' delete',
            del: 'delete',
        },
        operatoin: 'operatoin',
        status: 'status',
        request: 'request',
        response: 'response',
    },
    file: {
        dir: 'folder',
        upload: 'Upload',
        download: 'Download',
        fileName: 'file name',
        search: 'find',
        mode: 'permission',
        owner: 'owner',
        file: 'file',
        remoteFile: 'remote download',
        share: 'Share',
        sync: 'Data synchronization',
        size: 'size',
        updateTime: 'Modification time',
        open: 'open',
        rename: 'rename',
        role: 'authority',
        info: 'Properties',
        linkFile: 'soft link file',
        terminal: 'terminal',
        shareList: 'Share List',
        zip: 'compress',
        user: 'User',
        group: 'user group',
        path: 'path',
        public: 'public',
        setRole: 'Set permissions',
        link: 'Whether to link',
        rRole: 'read',
        wRole: 'Write',
        xRole: 'executable',
        name: 'name',
        compress: 'compress',
        deCompress: 'Decompress',
        compressType: 'compression format',
        compressDst: 'compression path',
        replace: 'Overwrite existing file',
        compressSuccess: 'Compression successful',
        deCompressSuccess: 'Decompression succeeded',
        deCompressDst: 'Decompression path',
        linkType: 'Link type',
        softLink: 'soft link',
        hardLink: 'hard link',
        linkPath: 'Link path',
        selectFile: 'Select file',
        downloadSuccess: 'Download successful',
        downloadUrl: 'download URL',
        downloadStart: 'Download started!',
        moveStart: 'Move start',
        move: 'Move',
        copy: 'Cpoy',
        calculate: 'Calculate',
        canNotDeCompress: 'Can not DeCompress this File',
        uploadSuccess: 'Upload Success!',
        downloadProcess: 'Download Process',
        downloading: 'Downloading...',
    },
    setting: {
        all: 'All',
        panel: 'Panel',
        emailHelper: 'For password retrieval',
        title: 'Panel alias',
        theme: 'Theme',
        dark: 'Dark',
        light: 'Light',
        language: 'Language',
        languageHelper:
            'By default, it follows the browser language. This parameter takes effect only on the current browser',
        sessionTimeout: 'Timeout',
        sessionTimeoutHelper:
            'If you do not operate the panel for more than {0} seconds, the panel automatically logs out',
        syncTime: 'Synchronization time',
        changePassword: 'Password change',
        oldPassword: 'Original password',
        newPassword: 'New password',
        retryPassword: 'Confirm password',

        backup: 'Backup',
        serverDisk: 'Server disks',
        OSS: 'Ali OSS',
        S3: 'Amazon S3',
        backupAccount: 'Backup account',
        loadBucket: 'Get bucket',
        accountName: 'Account name',
        accountKey: 'Account key',
        address: 'Address',
        port: 'Port',
        username: 'Username',
        password: 'Password',
        path: 'Path',

        safe: 'Safe',
        panelPort: 'Panel port',
        portHelper:
            'The recommended port range is 8888 to 65535. Note: If the server has a security group, permit the new port from the security group in advance',
        safeEntrance: 'Security entrance',
        safeEntranceHelper:
            'Panel management portal. You can log in to the panel only through a specified security portal, for example: onepanel',
        passwordTimeout: 'Expiration Time',
        timeoutHelper:
            '[ {0} days ] The panel password is about to expire. After the expiration, you need to reset the password',
        complexity: 'Complexity verification',
        complexityHelper:
            'The password must contain at least eight characters and contain at least three uppercase letters, lowercase letters, digits, and special characters',
        mfa: 'MFA',
        mfaHelper1: 'Download a MFA verification mobile app such as:',
        mfaHelper2: 'Scan the following QR code using the mobile app to obtain the 6-digit verification code',
        mfaHelper3: 'Enter six digits from the app',

        enableMonitor: 'Enable',
        storeDays: 'Expiration time (day)',
        cleanMonitor: 'Clearing monitoring records',

        message: 'Message',
        messageType: 'Message type',
        email: 'Email',
        wechat: 'WeChat',
        dingding: 'DingDing',
        closeMessage: 'Turning off Message Notification',
        emailServer: 'Service name',
        emailAddr: 'Service address',
        emailSMTP: 'SMTP code',
        secret: 'Secret',

        about: 'About',
        project: 'Project Address',
        issue: 'Feedback',
        chat: 'Community Discussion',
        star: 'Star',
        description: 'A modern Linux panel tool',
    },
};
