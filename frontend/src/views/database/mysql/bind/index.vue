<template>
    <DialogPro
        v-model="dialogVisible"
        :title="$t(props.viewOnly ? 'database.authorizedUsers' : 'database.authorizationManagement')"
        size="large"
    >
        <div v-loading="loading">
            <div v-if="!props.viewOnly" class="authorization-toolbar">
                <el-button type="primary" @click="openAddDialog">
                    {{ $t('database.addUserAuthorization') }}
                </el-button>
            </div>
            <el-table :data="authorizedUsers" :empty-text="$t('commons.msg.noneData')">
                <el-table-column :label="$t('commons.table.user')" show-overflow-tooltip min-width="180">
                    <template #default="{ row }">{{ row.username }}@{{ row.host }}</template>
                </el-table-column>
                <el-table-column :label="$t('commons.login.password')" prop="password" min-width="180">
                    <template #default="{ row }">
                        <span v-if="!row.password">-</span>
                        <div v-else class="password-cell">
                            <span v-if="!row.showPassword" class="password-text">**********</span>
                            <Tooltip v-else class="password-text" :islink="false" :text="row.password" />
                            <el-button
                                v-if="!row.showPassword"
                                link
                                @click="row.showPassword = true"
                                icon="View"
                                class="password-action"
                            />
                            <el-button
                                v-else
                                link
                                @click="row.showPassword = false"
                                icon="Hide"
                                class="password-action"
                            />
                            <CopyButton class="password-copy" :content="row.password" />
                        </div>
                    </template>
                </el-table-column>
                <el-table-column
                    :label="$t('commons.table.description')"
                    prop="description"
                    show-overflow-tooltip
                    min-width="120"
                />
                <el-table-column v-if="!props.viewOnly" :label="$t('commons.table.operate')" width="100" fixed="right">
                    <template #default="{ row }">
                        <el-button link type="primary" @click="revokeAuthorization(row)">
                            {{ $t('database.revokeAuthorization') }}
                        </el-button>
                    </template>
                </el-table-column>
            </el-table>
        </div>
        <template #footer>
            <el-button @click="dialogVisible = false" :disabled="loading">
                {{ $t('commons.button.close') }}
            </el-button>
        </template>
    </DialogPro>

    <DialogPro
        v-if="!props.viewOnly"
        v-model="addDialogVisible"
        :title="$t('database.addUserAuthorization')"
        size="small"
    >
        <el-form ref="formRef" :model="form" :rules="rules" label-position="top" v-loading="loading">
            <el-form-item :label="$t('commons.table.type')" prop="mode">
                <el-radio-group v-model="form.mode" @change="changeMode">
                    <el-radio-button value="select">
                        {{ $t('commons.button.select') }}
                    </el-radio-button>
                    <el-radio-button value="create">
                        {{ $t('commons.button.create') }}
                    </el-radio-button>
                </el-radio-group>
            </el-form-item>
            <el-form-item v-if="form.mode === 'select'" :label="$t('commons.login.username')" prop="userKey">
                <el-select
                    v-model="form.userKey"
                    filterable
                    popper-class="mysql-bind-user-select-popper"
                    @change="syncUser"
                >
                    <el-option
                        v-for="item in availableUsers"
                        :key="item.username + '@' + item.host"
                        class="user-select-option"
                        :label="item.username + ' - ' + permissionLabel(item.host)"
                        :value="item.username + '@' + item.host"
                    >
                        <div class="user-option-row">
                            <div class="user-option">
                                <span class="user-option-title">
                                    {{ item.username }} - {{ permissionLabel(item.host) }}
                                </span>
                                <span v-if="item.description" class="user-option-description">
                                    {{ item.description }}
                                </span>
                            </div>
                            <el-popover
                                :disabled="userDatabases(item).length === 0"
                                placement="right"
                                trigger="hover"
                                :width="220"
                            >
                                <template #reference>
                                    <el-button class="user-bound-button" size="small" @click.stop>
                                        {{ $t('database.authorizedDatabaseCount', [userDatabases(item).length]) }}
                                    </el-button>
                                </template>
                                <div class="user-bound-list">
                                    <el-tag
                                        v-for="database in userDatabases(item)"
                                        :key="item.username + item.host + database"
                                        size="small"
                                    >
                                        {{ database }}
                                    </el-tag>
                                </div>
                            </el-popover>
                        </div>
                    </el-option>
                </el-select>
            </el-form-item>
            <template v-else>
                <el-form-item :label="$t('commons.login.username')" prop="username">
                    <el-input v-model.trim="form.username" />
                </el-form-item>
                <el-form-item :label="$t('commons.login.password')" prop="password" required>
                    <el-input type="password" clearable show-password v-model.trim="form.password" />
                </el-form-item>
                <el-form-item :label="$t('database.permission')" prop="permission">
                    <el-select v-model="form.permission" @change="changePermission">
                        <el-option value="%" :label="$t('database.permissionAll')" />
                        <el-option value="ip" :label="$t('database.permissionForIP')" />
                    </el-select>
                </el-form-item>
                <el-form-item v-if="form.permission === 'ip'" :label="$t('database.permissionForIP')" prop="host">
                    <el-input v-model.trim="form.host" />
                    <span class="input-help">{{ $t('database.remoteHelper') }}</span>
                </el-form-item>
                <el-form-item :label="$t('commons.table.description')" prop="description">
                    <el-input type="textarea" clearable v-model="form.description" />
                </el-form-item>
            </template>
        </el-form>
        <template #footer>
            <el-button @click="addDialogVisible = false" :disabled="loading">
                {{ $t('commons.button.cancel') }}
            </el-button>
            <el-button type="primary" @click="submit" :disabled="loading">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DialogPro>
</template>

<script lang="ts" setup>
import { computed, reactive, ref } from 'vue';
import { ElMessageBox } from 'element-plus';
import Tooltip from '@/components/tooltip/index.vue';
import i18n from '@/lang';
import { Rules } from '@/global/form-rules';
import { MsgSuccess } from '@/utils/message';
import { Database } from '@/api/interface/database';
import {
    createMysqlUser,
    grantMysqlUser,
    revokeMysqlGrant,
    searchMysqlGrants,
    searchMysqlUsers,
} from '@/api/modules/database';

const emit = defineEmits<{ (e: 'search'): void }>();
const props = withDefaults(
    defineProps<{
        viewOnly?: boolean;
    }>(),
    {
        viewOnly: false,
    },
);

const dialogVisible = ref(false);
const addDialogVisible = ref(false);
const loading = ref(false);
const users = ref<Database.MysqlUser[]>([]);
const grants = ref<Database.MysqlGrant[]>([]);
const formRef = ref();
const form = reactive({
    database: '',
    db: '',
    mode: 'select',
    userKey: '',
    username: '',
    host: '%',
    permission: '%',
    password: '',
    description: '',
});
const checkMysqlHosts = (_rule: unknown, value: string, callback: (error?: Error) => void) => {
    const hosts = value.split(',').map((item) => item.trim());
    if (hosts.length === 0 || hosts.some((item) => !item)) {
        callback(new Error(i18n.global.t('commons.rule.requiredInput')));
        return;
    }
    callback();
};
const rules = reactive({
    userKey: [Rules.requiredSelect],
    username: [Rules.requiredInput, Rules.name],
    permission: [Rules.requiredSelect],
    host: [Rules.requiredInput, Rules.noSpace, Rules.illegal, { validator: checkMysqlHosts, trigger: 'blur' }],
    password: [Rules.requiredInput, Rules.noSpace, Rules.illegal],
});

interface DialogProps {
    database: string;
    db: string;
}

const grantKeys = computed(
    () =>
        new Set(
            grants.value.filter((item) => item.database === form.db).map((item) => `${item.username}@${item.host}`),
        ),
);
const authorizedUsers = computed(() =>
    users.value.filter((item) => !item.isDelete && grantKeys.value.has(`${item.username}@${item.host}`)),
);
const availableUsers = computed(() =>
    users.value.filter((item) => !item.isDelete && !grantKeys.value.has(`${item.username}@${item.host}`)),
);

const loadContext = async () => {
    const database = form.database;
    if (!database) {
        users.value = [];
        grants.value = [];
        return;
    }
    const [userRes, grantRes] = await Promise.all([searchMysqlUsers({ database }), searchMysqlGrants({ database })]);
    users.value = userRes.data || [];
    grants.value = grantRes.data || [];
};

const permissionLabel = (host: string) => {
    return host === '%' ? i18n.global.t('database.permissionAll') : host;
};

const userDatabases = (user: Database.MysqlUser) => {
    return Array.from(
        new Set(
            grants.value
                .filter((item) => item.username === user.username && item.host === user.host)
                .map((item) => item.database),
        ),
    );
};

const changePermission = () => {
    if (form.permission === '%') {
        form.host = '%';
    }
};

const changeMode = () => {
    if (form.mode === 'create') {
        form.username = '';
        form.host = '%';
        form.permission = '%';
        form.password = '';
        form.description = '';
        return;
    }
    const user = availableUsers.value[0];
    form.userKey = user ? `${user.username}@${user.host}` : '';
    syncUser();
};

const openAddDialog = () => {
    form.mode = availableUsers.value.length ? 'select' : 'create';
    form.password = '';
    form.description = '';
    changeMode();
    addDialogVisible.value = true;
};

const acceptParams = async (params: DialogProps) => {
    if (!params.database || !params.db) {
        return;
    }
    form.database = params.database;
    form.db = params.db;
    dialogVisible.value = true;
    loading.value = true;
    try {
        await loadContext();
    } finally {
        loading.value = false;
    }
};

const syncUser = () => {
    const item = availableUsers.value.find((item) => `${item.username}@${item.host}` === form.userKey);
    if (item) {
        form.username = item.username;
        form.host = item.host;
    }
};

const submit = async () => {
    if (!formRef.value) return;
    await formRef.value.validate(async (valid: boolean) => {
        if (!valid) return;
        loading.value = true;
        try {
            if (form.mode === 'create') {
                await createMysqlUser({
                    database: form.database,
                    username: form.username,
                    host: form.host,
                    password: form.password,
                    description: form.description,
                });
            }
            await grantMysqlUser({
                database: form.database,
                db: form.db,
                username: form.username,
                host: form.host,
            });
            await loadContext();
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            addDialogVisible.value = false;
            emit('search');
        } finally {
            loading.value = false;
        }
    });
};

const revokeAuthorization = (user: Database.MysqlUser) => {
    ElMessageBox.confirm(
        i18n.global.t('database.revokeAuthorizationHelper', [`${user.username}@${user.host}`, form.db]),
        i18n.global.t('commons.msg.infoTitle'),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'warning',
        },
    ).then(async () => {
        loading.value = true;
        try {
            await revokeMysqlGrant({
                database: form.database,
                db: form.db,
                username: user.username,
                host: user.host,
            });
            await loadContext();
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            emit('search');
        } finally {
            loading.value = false;
        }
    });
};

defineExpose({
    acceptParams,
});
</script>

<style lang="scss" scoped>
.authorization-toolbar {
    display: flex;
    margin-bottom: 12px;
}
.password-cell {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    max-width: 100%;
}
.password-text {
    overflow: hidden;
    max-width: 96px;
    text-overflow: ellipsis;
    white-space: nowrap;
}
.password-action {
    min-height: 20px;
    padding: 0;
}
.password-copy {
    margin-left: -2px;
}
:deep(.user-select-option) {
    height: auto;
    min-height: 48px;
}
.user-option-row {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
}
.user-option {
    display: flex;
    flex-direction: column;
    justify-content: center;
    flex: 1;
    min-width: 0;
    min-height: 40px;
    padding: 4px 0;
    line-height: 18px;
}
.user-option-title {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
.user-option-description {
    margin-top: 3px;
    color: var(--el-text-color-secondary);
    font-size: 12px;
    line-height: 16px;
    white-space: normal;
    word-break: break-word;
}
.user-bound-button {
    flex-shrink: 0;
}
</style>

<style lang="scss">
.mysql-bind-user-select-popper {
    min-width: 360px;

    .user-select-option {
        height: auto;
        min-height: 48px;
        padding-top: 4px;
        padding-bottom: 4px;
    }
}
.user-bound-list {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
}
</style>
