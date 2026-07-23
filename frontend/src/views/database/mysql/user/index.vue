<template>
    <DrawerPro v-model="drawerVisible" :header="$t('commons.table.user')" size="60%">
        <div class="drawer-toolbar">
            <el-button type="primary" @click="openUserDialog()">
                {{ $t('commons.button.create') }}
            </el-button>
        </div>
        <ComplexTable :data="users" :heightDiff="260">
            <el-table-column :label="$t('commons.table.user')" show-overflow-tooltip prop="username" min-width="180">
                <template #default="{ row }">
                    <span>{{ row.username }}@{{ row.host }}</span>
                    <el-tag v-if="row.isDelete" round type="info" class="ml-1" size="small">
                        {{ $t('database.isDelete') }}
                    </el-tag>
                </template>
            </el-table-column>
            <el-table-column :label="$t('commons.login.password')" prop="password" min-width="120">
                <template #default="{ row }">
                    <span v-if="row.isDelete">-</span>
                    <div v-else-if="!row.password" class="password-cell">
                        <el-button link type="primary" @click="openSupplementPasswordDialog(row)">
                            {{ $t('database.supplementPassword') }}
                        </el-button>
                    </div>
                    <div class="password-cell" v-else>
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
                            v-if="row.showPassword"
                            link
                            @click="row.showPassword = false"
                            icon="Hide"
                            class="password-action"
                        />
                        <CopyButton class="password-copy" :content="row.password" />
                    </div>
                </template>
            </el-table-column>
            <el-table-column :label="$t('menu.database')" min-width="160">
                <template #default="{ row }">
                    <span v-if="userDatabases(row).length === 0">-</span>
                    <div v-else class="bind-db-list">
                        <el-tooltip
                            v-for="item in userDatabases(row).slice(0, 1)"
                            :key="row.username + '@' + row.host + item"
                            :content="item"
                            placement="top"
                        >
                            <el-tag size="small">{{ item }}</el-tag>
                        </el-tooltip>
                        <el-popover v-if="userDatabases(row).length > 1" placement="right" trigger="click" :width="260">
                            <template #reference>
                                <el-tag class="cursor-pointer" type="info" size="small">
                                    +{{ userDatabases(row).length - 1 }}
                                </el-tag>
                            </template>
                            <div class="bind-db-popover">
                                <el-tooltip
                                    v-for="item in userDatabases(row)"
                                    :key="row.username + '@' + row.host + item"
                                    :content="item"
                                    placement="top"
                                >
                                    <el-tag size="small">{{ item }}</el-tag>
                                </el-tooltip>
                            </div>
                        </el-popover>
                    </div>
                </template>
            </el-table-column>
            <el-table-column
                :label="$t('commons.table.description')"
                prop="description"
                show-overflow-tooltip
                min-width="120"
            />
            <fu-table-operations
                :width="120"
                :buttons="userButtons"
                :label="$t('commons.table.operate')"
                fixed="right"
                fix
            />
        </ComplexTable>
    </DrawerPro>

    <DialogPro v-model="userDialogVisible" :title="userDialogTitle" size="small">
        <el-form ref="userFormRef" :model="userForm" :rules="userRules" label-position="top" v-loading="loading">
            <el-form-item :label="$t('commons.login.username')" prop="username">
                <el-input v-model.trim="userForm.username" :disabled="userDialogMode === 'edit'" />
            </el-form-item>
            <el-form-item :label="$t('database.permission')" prop="permission">
                <el-select v-model="userForm.permission" @change="changePermission">
                    <el-option value="%" :label="$t('database.permissionAll')" />
                    <el-option value="ip" :label="$t('database.permissionForIP')" />
                </el-select>
            </el-form-item>
            <el-form-item v-if="userForm.permission === 'ip'" prop="host">
                <el-input v-model.trim="userForm.host" />
                <span v-if="userDialogMode === 'create'" class="input-help">
                    {{ $t('database.remoteHelper') }}
                </span>
            </el-form-item>
            <el-form-item :label="$t('commons.login.password')" prop="password" :required="userDialogMode === 'create'">
                <el-input type="password" clearable show-password v-model.trim="userForm.password" />
                <span v-if="userDialogMode === 'edit'" class="input-help">
                    {{ $t('setting.passwordEmptyTip') }}
                </span>
            </el-form-item>
            <el-form-item :label="$t('commons.table.description')" prop="description">
                <el-input type="textarea" clearable v-model="userForm.description" />
            </el-form-item>
            <el-form-item :label="$t('menu.database')" prop="dbs">
                <el-select v-model="userForm.dbs" filterable multiple collapse-tags-tooltip>
                    <el-option v-for="item in databases" :key="item.name" :label="item.name" :value="item.name" />
                </el-select>
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button @click="userDialogVisible = false" :disabled="loading">
                {{ $t('commons.button.cancel') }}
            </el-button>
            <el-button type="primary" @click="submitUser()" :disabled="loading">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DialogPro>

    <DialogPro v-model="supplementDialogVisible" :title="$t('database.supplementPassword')" size="small">
        <el-form
            ref="supplementFormRef"
            :model="supplementForm"
            :rules="supplementRules"
            label-position="top"
            v-loading="loading"
        >
            <el-form-item :label="$t('commons.login.username')">
                <el-input :model-value="`${supplementForm.username}@${supplementForm.host}`" disabled />
            </el-form-item>
            <el-form-item :label="$t('commons.login.password')" prop="password">
                <el-input type="password" clearable show-password v-model.trim="supplementForm.password" />
                <span class="input-help">{{ $t('database.supplementPasswordHelper') }}</span>
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button @click="supplementDialogVisible = false" :disabled="loading">
                {{ $t('commons.button.cancel') }}
            </el-button>
            <el-button type="primary" @click="submitSupplementPassword" :disabled="loading">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DialogPro>

    <DialogPro
        v-model="deleteDialogVisible"
        :title="$t('commons.button.delete') + ' - ' + deleteUserIdentity"
        size="small"
    >
        <div v-loading="loading">
            <div class="delete-user-title">{{ deleteUserIdentity }} {{ $t('database.userBoundDatabases') }}</div>
            <div class="delete-user-section">
                <div v-if="deleteUserDbs.length === 0" class="delete-user-empty">-</div>
                <div v-else class="bind-db-list delete-user-dbs">
                    <el-tag v-for="item in deleteUserDbs" :key="item" size="small">
                        {{ item }}
                    </el-tag>
                </div>
            </div>
            <div class="delete-user-section">
                <div>
                    <span style="font-size: 12px">{{ $t('database.delete') }}</span>
                    <span style="font-size: 12px; color: red; font-weight: 500">
                        {{ deleteUserIdentity }}
                    </span>
                    <span style="font-size: 12px">{{ $t('database.deleteUserHelper') }}</span>
                </div>
                <el-input v-model="deleteConfirmInput" :placeholder="deleteUserIdentity" />
            </div>
        </div>
        <template #footer>
            <el-button @click="deleteDialogVisible = false" :disabled="loading">
                {{ $t('commons.button.cancel') }}
            </el-button>
            <el-button
                type="primary"
                @click="submitDeleteUser()"
                :disabled="deleteConfirmInput !== deleteUserIdentity || loading"
            >
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DialogPro>
</template>

<script lang="ts" setup>
import { computed, reactive, ref } from 'vue';
import Tooltip from '@/components/tooltip/index.vue';
import i18n from '@/lang';
import { Rules } from '@/global/form-rules';
import { MsgSuccess } from '@/utils/message';
import { Database } from '@/api/interface/database';
import {
    createMysqlUser,
    deleteMysqlUser,
    grantMysqlUser,
    revokeMysqlGrant,
    searchMysqlDBs,
    searchMysqlGrants,
    searchMysqlUsers,
    saveMysqlUserPassword,
    updateMysqlUser,
    updateMysqlUserPassword,
} from '@/api/modules/database';

const emit = defineEmits<{ (e: 'search'): void }>();

const drawerVisible = ref(false);
const loading = ref(false);
const database = ref('');
const users = ref<Database.MysqlUser[]>([]);
const grants = ref<Database.MysqlGrant[]>([]);
const databases = ref<Database.MysqlDBInfo[]>([]);

const userDialogVisible = ref(false);
const userDialogMode = ref<'create' | 'edit'>('create');
const userDialogTitle = computed(() =>
    userDialogMode.value === 'create' ? i18n.global.t('commons.table.user') : i18n.global.t('commons.button.edit'),
);
const userFormRef = ref();
const userForm = reactive({
    username: '',
    host: '%',
    permission: '%',
    password: '',
    description: '',
    dbs: [] as string[],
});
const originalDbs = ref<string[]>([]);
const originalHost = ref('');
const originalDescription = ref('');
const deleteDialogVisible = ref(false);
const deleteConfirmInput = ref('');
const deleteUser = reactive({
    username: '',
    host: '',
    isDelete: false,
});
const deleteUserIdentity = computed(() => `${deleteUser.username}@${deleteUser.host}`);
const deleteUserDbs = ref<string[]>([]);
const supplementDialogVisible = ref(false);
const supplementFormRef = ref();
const supplementForm = reactive({
    username: '',
    host: '',
    password: '',
});
const supplementRules = reactive({
    password: [Rules.requiredInput, Rules.noSpace, Rules.illegal],
});
const checkPassword = (_rule: unknown, value: string, callback: (error?: Error) => void) => {
    if (!value && userDialogMode.value === 'edit') {
        callback();
        return;
    }
    if (!value) {
        callback(new Error(i18n.global.t('commons.rule.requiredInput')));
        return;
    }
    if (value.indexOf(' ') !== -1) {
        callback(new Error(i18n.global.t('setting.noSpace')));
        return;
    }
    if (/[&|;'`$()><]/.test(value)) {
        callback(new Error(i18n.global.t('commons.rule.illegalInput')));
        return;
    }
    callback();
};
const checkUserHosts = (_rule: unknown, value: string, callback: (error?: Error) => void) => {
    const hosts = value.split(',').map((item) => item.trim());
    if (hosts.length === 0 || hosts.some((item) => !item)) {
        callback(new Error(i18n.global.t('commons.rule.requiredInput')));
        return;
    }
    if (userDialogMode.value === 'edit' && hosts.length > 1) {
        callback(new Error(i18n.global.t('commons.rule.illegalInput')));
        return;
    }
    callback();
};
const userRules = reactive({
    username: [Rules.requiredInput, Rules.name],
    permission: [Rules.requiredSelect],
    host: [Rules.requiredInput, Rules.noSpace, Rules.illegal, { validator: checkUserHosts, trigger: 'blur' }],
    password: [{ validator: checkPassword, trigger: 'blur' }],
});

interface DialogProps {
    database: string;
}

const acceptParams = async (params: DialogProps) => {
    if (!params.database) {
        return;
    }
    database.value = params.database;
    drawerVisible.value = true;
    await loadContext();
};

const loadUsers = async () => {
    const databaseName = database.value;
    if (!databaseName) {
        users.value = [];
        return;
    }
    const res = await searchMysqlUsers({ database: databaseName });
    users.value = res.data || [];
};

const loadGrants = async () => {
    const databaseName = database.value;
    if (!databaseName) {
        grants.value = [];
        return;
    }
    const res = await searchMysqlGrants({ database: databaseName });
    grants.value = res.data || [];
};

const loadDatabases = async () => {
    const res = await searchMysqlDBs({
        page: 1,
        pageSize: 10000,
        info: '',
        database: database.value,
        orderBy: 'createdAt',
        order: 'null',
    });
    databases.value = res.data.items || [];
};

const loadContext = async () => {
    if (!drawerVisible.value || !database.value) {
        return;
    }
    await Promise.all([loadUsers(), loadGrants(), loadDatabases()]);
};

const userDatabaseMap = computed(() => {
    const databaseSets = new Map<string, Set<string>>();
    for (const grant of grants.value) {
        const key = `${grant.username}@${grant.host}`;
        let databases = databaseSets.get(key);
        if (!databases) {
            databases = new Set<string>();
            databaseSets.set(key, databases);
        }
        databases.add(grant.database);
    }

    const result = new Map<string, string[]>();
    for (const [key, databases] of databaseSets) {
        result.set(key, Array.from(databases));
    }
    return result;
});

const userDatabases = (row: Database.MysqlUser) => {
    return userDatabaseMap.value.get(`${row.username}@${row.host}`) || [];
};

const changePermission = () => {
    if (userForm.permission === '%') {
        userForm.host = '%';
    }
};

const openDeleteUserDialog = (row: Database.MysqlUser) => {
    deleteUser.username = row.username;
    deleteUser.host = row.host;
    deleteUser.isDelete = row.isDelete;
    deleteUserDbs.value = userDatabases(row);
    deleteConfirmInput.value = '';
    deleteDialogVisible.value = true;
};

const openSupplementPasswordDialog = (row: Database.MysqlUser) => {
    supplementForm.username = row.username;
    supplementForm.host = row.host;
    supplementForm.password = '';
    supplementDialogVisible.value = true;
};

const submitSupplementPassword = async () => {
    if (!supplementFormRef.value) return;
    await supplementFormRef.value.validate(async (valid: boolean) => {
        if (!valid) return;
        loading.value = true;
        try {
            await saveMysqlUserPassword({
                database: database.value,
                username: supplementForm.username,
                host: supplementForm.host,
                password: supplementForm.password,
            });
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            supplementDialogVisible.value = false;
            await loadUsers();
        } finally {
            loading.value = false;
        }
    });
};

const openUserDialog = () => {
    userDialogMode.value = 'create';
    userForm.username = '';
    userForm.host = '%';
    userForm.permission = '%';
    userForm.password = '';
    userForm.description = '';
    userForm.dbs = [];
    originalDbs.value = [];
    originalHost.value = '%';
    originalDescription.value = '';
    userDialogVisible.value = true;
};

const openUserEditDialog = async (row: Database.MysqlUser) => {
    if (!databases.value.length) {
        await loadDatabases();
    }
    userDialogMode.value = 'edit';
    userForm.username = row.username;
    userForm.host = row.host;
    userForm.permission = row.host === '%' ? '%' : 'ip';
    userForm.password = '';
    userForm.description = row.description || '';
    userForm.dbs = userDatabases(row);
    originalDbs.value = [...userForm.dbs];
    originalHost.value = row.host;
    originalDescription.value = row.description || '';
    userDialogVisible.value = true;
};

const submitUser = async () => {
    if (!userFormRef.value) return;
    await userFormRef.value.validate(async (valid: boolean) => {
        if (!valid) return;
        loading.value = true;
        if (userDialogMode.value === 'edit') {
            try {
                const selectedDbs = userForm.dbs || [];
                const addDbs = selectedDbs.filter((item) => !originalDbs.value.includes(item));
                const removeDbs = originalDbs.value.filter((item) => !selectedDbs.includes(item));
                if (userForm.host !== originalHost.value || userForm.description !== originalDescription.value) {
                    await updateMysqlUser({
                        database: database.value,
                        username: userForm.username,
                        host: originalHost.value,
                        newHost: userForm.host,
                        description: userForm.description,
                    });
                }
                if (userForm.password) {
                    await updateMysqlUserPassword({
                        database: database.value,
                        username: userForm.username,
                        host: userForm.host,
                        password: userForm.password,
                    });
                }
                for (const item of addDbs) {
                    await grantMysqlUser({
                        database: database.value,
                        db: item,
                        username: userForm.username,
                        host: userForm.host,
                    });
                }
                for (const item of removeDbs) {
                    await revokeMysqlGrant({
                        database: database.value,
                        db: item,
                        username: userForm.username,
                        host: userForm.host,
                    });
                }
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                userDialogVisible.value = false;
                emit('search');
            } finally {
                await loadContext();
                loading.value = false;
            }
            return;
        }
        try {
            await createMysqlUser({
                database: database.value,
                username: userForm.username,
                host: userForm.host,
                password: userForm.password,
                description: userForm.description,
                dbs: userForm.dbs,
            });
            await loadContext();
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            userDialogVisible.value = false;
            emit('search');
        } finally {
            loading.value = false;
        }
    });
};

const userButtons = [
    {
        label: i18n.global.t('commons.button.edit'),
        permission: true,
        disabled: (row: Database.MysqlUser) => row.isDelete,
        click: (row: Database.MysqlUser) => {
            openUserEditDialog(row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        permission: true,
        click: (row: Database.MysqlUser) => {
            openDeleteUserDialog(row);
        },
    },
];

const submitDeleteUser = async () => {
    loading.value = true;
    await deleteMysqlUser({
        database: database.value,
        username: deleteUser.username,
        host: deleteUser.host,
    })
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.deleteSuccess'));
            deleteDialogVisible.value = false;
            loadContext();
            emit('search');
        })
        .finally(() => {
            loading.value = false;
        });
};

defineExpose({
    acceptParams,
});
</script>

<style lang="scss" scoped>
.drawer-toolbar {
    display: flex;
    gap: 8px;
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
.bind-db-list {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
}
.bind-db-popover {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 8px;

    :deep(.el-tag) {
        width: 100%;
        justify-content: flex-start;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }
}
.delete-user-title {
    margin-bottom: 12px;
    font-size: 14px;
    font-weight: 500;
    line-height: 20px;
}
.delete-user-section {
    margin-top: 12px;
}
.delete-user-empty {
    margin-top: 6px;
}
.delete-user-dbs {
    margin-top: 6px;
}
</style>
