export default {
    commons: {
        button: {
            create: 'Create',
            add: 'Add',
            delete: 'Delete',
            edit: 'Edit',
            confirm: 'Confirm',
            cancel: 'Cancel',
            reset: 'Reset',
            login: 'Login',
            conn: 'Connect',
        },
        table: {
            name: 'Name',
            createdAt: 'Creation Time',
            updatedAt: 'Update Time',
            operate: 'Operations',
            message: 'message',
        },
        msg: {
            delete: 'This operation cannot be rolled back. Do you want to continue',
            title: 'Delete',
            deleteSuccess: 'Delete Success',
            loginSuccess: 'Login Success',
            requestTimeout: 'The request timed out, please try again later',
            deleteTitle: 'Delete',
            operationSuccess: 'Successful operation',
            infoTitle: 'Hint',
            sureLogOut: 'Are you sure you want to log out?',
            createSuccess: 'Create Success',
            updateSuccess: 'Update Success',
        },
        login: {
            captchaHelper: 'Please enter the verification code',
        },
        rule: {
            username: 'Please enter a username',
            password: 'Please enter a password',
            requiredInput: 'Please enter the required fields',
            requiredSelect: 'Please select the required fields',
            commonName: 'Support English, Chinese, numbers, .-_, length 1-30',
            email: 'Email format error',
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
    },
    business: {
        user: {
            username: 'Username',
            email: 'Email',
            password: 'Password',
        },
    },
    menu: {
        home: 'Dashboard',
        demo: 'Demo',
        terminal: 'Terminal',
        operations: 'Operation logs',
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
        componentSize: 'Component size',
        language: 'Language',
        theme: 'theme',
        layoutConfig: 'Layout config',
        primary: 'primary',
        darkMode: 'Dark Mode',
        greyMode: 'Grey mode',
        weakMode: 'Weak mode',
        fullScreen: 'Full Screen',
        exitFullScreen: 'Exit Full Screen',
        personalData: 'Personal Data',
        changePassword: 'Change Password',
        logout: 'Logout',
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
};
