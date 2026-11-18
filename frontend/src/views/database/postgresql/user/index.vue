<template>
    <DrawerPro
        v-model="visible"
        :header="name ? $t('database.authorizationManagement') : $t('commons.table.user')"
        :resource="name || database"
        size="60%"
    >
        <div v-loading="loading">
            <el-alert
                :title="$t('database.pgAuthorizationHelper')"
                type="info"
                :closable="false"
                class="authorization-hint"
            />
            <div class="drawer-toolbar">
                <el-button v-if="!name" v-permission type="primary" @click="openEdit()">
                    {{ $t('commons.button.create') }}
                </el-button>
                <el-button v-else v-permission type="primary" @click="openAddAuthorization">
                    {{ $t('database.addUserAuthorization') }}
                </el-button>
            </div>
            <ComplexTable :data="displayed" :heightDiff="260">
                <el-table-column :label="$t('commons.table.user')" prop="username" min-width="180">
                    <template #default="{ row }">
                        <div class="user-name-row">
                            <el-tooltip :content="row.username" v-bind="TABLE_TOOLTIP_OPTIONS">
                                <span class="user-name-text">{{ row.username }}</span>
                            </el-tooltip>
                            <el-tag v-if="row.superUser" class="superuser-tag" type="warning" size="small">
                                {{ $t('database.pgSuperUser') }}
                            </el-tag>
                        </div>
                    </template>
                </el-table-column>
                <el-table-column :label="$t('commons.login.password')" prop="password" min-width="120">
                    <template #default="{ row }">
                        <div v-if="row.password" class="password-cell">
                            <span v-if="!row.showPassword" class="password-text">**********</span>
                            <Tooltip v-else class="password-text" :islink="false" :text="row.password" />
                            <el-button
                                link
                                :icon="row.showPassword ? 'Hide' : 'View'"
                                class="password-action"
                                @click="row.showPassword = !row.showPassword"
                            />
                            <CopyButton class="password-copy" :content="row.password" />
                        </div>
                        <div v-else-if="!row.protected" class="password-cell">
                            <el-button v-permission link type="primary" @click="openSupplementPassword(row)">
                                {{ $t('database.supplementPassword') }}
                            </el-button>
                        </div>
                        <span v-else>-</span>
                    </template>
                </el-table-column>
                <el-table-column :label="$t('menu.database')" min-width="160">
                    <template #default="{ row }">
                        <el-tag v-if="row.superUser" size="small">{{ $t('commons.table.all') }}</el-tag>
                        <span v-else-if="!row.databases.length">-</span>
                        <div v-else class="bind-db-list">
                            <el-tooltip
                                v-for="db in row.databases.slice(0, 1)"
                                :key="db"
                                :content="db"
                                v-bind="TABLE_TOOLTIP_OPTIONS"
                            >
                                <el-tag size="small">{{ db }}</el-tag>
                            </el-tooltip>
                            <el-popover v-if="row.databases.length > 1" placement="right" trigger="click" :width="260">
                                <template #reference>
                                    <el-tag class="cursor-pointer" type="info" size="small">
                                        +{{ row.databases.length - 1 }}
                                    </el-tag>
                                </template>
                                <div class="bind-db-popover">
                                    <el-tooltip
                                        v-for="db in row.databases"
                                        :key="db"
                                        :content="db"
                                        v-bind="TABLE_TOOLTIP_OPTIONS"
                                    >
                                        <el-tag size="small">{{ db }}</el-tag>
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
        </div>
    </DrawerPro>
    <DialogPro v-model="editing" :title="dialogTitle" size="small">
        <el-form ref="formRef" v-loading="saving" :model="form" :rules="rules" label-position="top">
            <el-form-item v-if="name && (mode === 'select' || mode === 'create')" :label="$t('commons.table.type')">
                <el-radio-group v-model="mode" @change="changeAuthorizationMode">
                    <el-radio-button value="select">{{ $t('commons.button.select') }}</el-radio-button>
                    <el-radio-button value="create">{{ $t('commons.button.create') }}</el-radio-button>
                </el-radio-group>
            </el-form-item>
            <el-form-item v-if="mode === 'select'" :label="$t('commons.login.username')" prop="username">
                <el-select
                    v-model="form.username"
                    filterable
                    :fit-input-width="true"
                    popper-class="postgresql-bind-user-select-popper"
                >
                    <el-option
                        v-for="user in available"
                        :key="user.username"
                        :value="user.username"
                        :label="user.username"
                        class="user-select-option"
                    >
                        <div class="user-option-row">
                            <div class="user-option">
                                <el-tooltip :content="user.username" v-bind="TABLE_TOOLTIP_OPTIONS">
                                    <span class="user-option-title">{{ user.username }}</span>
                                </el-tooltip>
                                <span v-if="user.description" class="user-option-description">
                                    {{ user.description }}
                                </span>
                            </div>
                            <el-popover
                                :disabled="!user.databases.length"
                                placement="right"
                                trigger="hover"
                                :width="220"
                            >
                                <template #reference>
                                    <el-button class="user-bound-button" size="small" @click.stop>
                                        {{ $t('database.authorizedDatabaseCount', [user.databases.length]) }}
                                    </el-button>
                                </template>
                                <div class="user-bound-list">
                                    <el-tag v-for="db in user.databases" :key="db" size="small">{{ db }}</el-tag>
                                </div>
                            </el-popover>
                        </div>
                    </el-option>
                </el-select>
            </el-form-item>
            <template v-else>
                <el-form-item :label="$t('commons.login.username')" prop="username">
                    <el-input v-model.trim="form.username" :disabled="mode !== 'create'" />
                </el-form-item>
                <el-form-item v-if="mode !== 'supplement'" :label="$t('database.permission')">
                    <el-checkbox v-model="form.superUser">{{ $t('database.pgSuperUser') }}</el-checkbox>
                    <span class="input-help">{{ $t('database.pgAuthorizationHelper') }}</span>
                </el-form-item>
                <el-form-item :label="$t('commons.login.password')" prop="password" :required="mode !== 'edit'">
                    <el-input v-model.trim="form.password" type="password" clearable show-password />
                    <span class="input-help">{{ $t('commons.rule.illegalChar') }}</span>
                    <span v-if="mode === 'edit'" class="input-help">{{ $t('setting.passwordEmptyTip') }}</span>
                    <span v-if="mode === 'supplement'" class="input-help">
                        {{ $t('database.supplementPasswordHelper') }}
                    </span>
                </el-form-item>
            </template>
            <template v-if="mode === 'create' || mode === 'edit'">
                <el-form-item :label="$t('commons.table.description')" prop="description">
                    <el-input v-model="form.description" type="textarea" clearable />
                </el-form-item>
            </template>
        </el-form>
        <template #footer>
            <el-button :disabled="saving" @click="editing = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button type="primary" :disabled="saving" @click="save">{{ $t('commons.button.confirm') }}</el-button>
        </template>
    </DialogPro>

    <DialogPro
        v-model="deleteDialogVisible"
        :title="$t('commons.button.delete') + ' - ' + deleteUser.username"
        size="small"
    >
        <div v-loading="deleting">
            <div class="delete-user-title">{{ deleteUser.username }} {{ $t('database.userBoundDatabases') }}</div>
            <div class="delete-user-section">
                <el-tag v-if="deleteUser.superUser" size="small">{{ $t('commons.table.all') }}</el-tag>
                <div v-else-if="deleteUser.databases.length === 0" class="delete-user-empty">-</div>
                <div v-else class="bind-db-list delete-user-dbs">
                    <el-tag v-for="item in deleteUser.databases" :key="item" size="small">
                        {{ item }}
                    </el-tag>
                </div>
            </div>
            <div class="delete-user-section">
                <div>
                    <span style="font-size: 12px">{{ $t('database.delete') }}</span>
                    <span style="font-size: 12px; color: red; font-weight: 500">
                        {{ deleteUser.username }}
                    </span>
                    <span style="font-size: 12px">{{ $t('database.deleteUserHelper') }}</span>
                </div>
                <el-input v-model="deleteConfirmInput" :placeholder="deleteUser.username" />
            </div>
        </div>
        <template #footer>
            <el-button @click="deleteDialogVisible = false" :disabled="deleting">
                {{ $t('commons.button.cancel') }}
            </el-button>
            <el-button
                type="primary"
                @click="remove()"
                :disabled="deleteConfirmInput !== deleteUser.username || deleting"
            >
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DialogPro>
</template>
<script setup lang="ts">
import { computed, reactive, ref } from 'vue';
import Tooltip from '@/components/tooltip/index.vue';
import { TABLE_TOOLTIP_OPTIONS } from '@/components/table/shared';
import { ElMessageBox, type FormInstance } from 'element-plus';
import { Database } from '@/api/interface/database';
import {
    searchPostgresqlUsers,
    createPostgresqlUser,
    updatePostgresqlUser,
    changePostgresqlUserPassword,
    savePostgresqlUserPassword,
    deletePostgresqlUser,
    grantPostgresqlUser,
    revokePostgresqlUser,
} from '@/api/modules/database';
import { Rules } from '@/global/form-rules';
import { MsgSuccess } from '@/utils/message';
import i18n from '@/lang';

const visible = ref(false);
const loading = ref(false);
const saving = ref(false);
const database = ref('');
const name = ref('');
const users = ref<Array<Database.PostgresqlUser & { showPassword?: boolean }>>([]);
const displayed = computed(() =>
    users.value.filter((user) => !name.value || user.superUser || user.databases.includes(name.value)),
);
const available = computed(() =>
    users.value.filter((user) => !user.superUser && !user.protected && !user.databases.includes(name.value)),
);
const editing = ref(false);
const mode = ref<'create' | 'edit' | 'select' | 'supplement'>('create');
const dialogTitle = computed(() => {
    if (name.value && (mode.value === 'create' || mode.value === 'select')) {
        return i18n.global.t('database.addUserAuthorization');
    }
    if (mode.value === 'supplement') return i18n.global.t('database.supplementPassword');
    return i18n.global.t(mode.value === 'create' ? 'commons.table.user' : 'commons.button.edit');
});
const deleteDialogVisible = ref(false);
const deleting = ref(false);
const deleteConfirmInput = ref('');
const deleteUser = reactive({ username: '', databases: [] as string[], superUser: false });
const formRef = ref<FormInstance>();
const form = reactive({ username: '', password: '', superUser: false, description: '' });
const rules = computed(() => ({
    username:
        mode.value === 'create'
            ? [Rules.requiredInput, Rules.name]
            : mode.value === 'select'
              ? [Rules.requiredSelect]
              : [Rules.requiredInput],
    password: mode.value === 'edit' ? [Rules.databasePassword] : [Rules.requiredInput, Rules.databasePassword],
}));
const emit = defineEmits<{ (e: 'search'): void }>();

const load = async () => {
    loading.value = true;
    try {
        const res = await searchPostgresqlUsers(database.value);
        users.value = res.data || [];
    } catch {
        return;
    } finally {
        loading.value = false;
    }
};
const acceptParams = async (params: { database: string; name?: string }) => {
    database.value = params.database;
    name.value = params.name || '';
    users.value = [];
    visible.value = true;
    await load();
};
const openEdit = (user?: Database.PostgresqlUser) => {
    mode.value = user ? 'edit' : 'create';
    Object.assign(form, {
        username: user?.username || '',
        password: '',
        superUser: user?.superUser || false,
        description: user?.description || '',
    });
    editing.value = true;
    formRef.value?.clearValidate();
};
const openAddAuthorization = () => {
    openEdit();
    mode.value = 'select';
};
const changeAuthorizationMode = () => {
    Object.assign(form, { username: '', password: '', superUser: false, description: '' });
    formRef.value?.clearValidate();
};
const openDeleteUserDialog = (user: Database.PostgresqlUser) => {
    Object.assign(deleteUser, { username: user.username, databases: user.databases, superUser: user.superUser });
    deleteConfirmInput.value = '';
    deleteDialogVisible.value = true;
};
const openSupplementPassword = (user: Database.PostgresqlUser) => {
    openEdit(user);
    mode.value = 'supplement';
};
const save = async () => {
    if (!(await formRef.value?.validate().catch(() => false))) return;
    saving.value = true;
    try {
        const params = { database: database.value, ...form };
        if (mode.value === 'create') {
            await createPostgresqlUser(params);
            mode.value = name.value && !form.superUser ? 'select' : 'edit';
            form.password = '';
            if (name.value && !form.superUser) {
                await grantPostgresqlUser({ database: database.value, name: name.value, username: form.username });
            }
        } else if (mode.value === 'supplement') {
            await savePostgresqlUserPassword({
                database: database.value,
                username: form.username,
                password: form.password,
            });
        } else if (mode.value === 'select') {
            await grantPostgresqlUser({ database: database.value, name: name.value, username: form.username });
        } else {
            await updatePostgresqlUser({
                database: database.value,
                username: form.username,
                superUser: form.superUser,
                description: form.description,
            });
            if (form.password) {
                await changePostgresqlUserPassword({
                    database: database.value,
                    username: form.username,
                    password: form.password,
                });
            }
        }
        editing.value = false;
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        emit('search');
    } catch {
        return;
    } finally {
        saving.value = false;
        await load();
    }
};
const revoke = async (user: Database.PostgresqlUser) => {
    const confirmed = await ElMessageBox.confirm(
        `${i18n.global.t('database.revokeAuthorizationHelper', [user.username, name.value])} ${i18n.global.t('database.pgRevokeAuthorizationHelper')}`,
        i18n.global.t('commons.msg.infoTitle'),
        {
            type: 'warning',
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
        },
    ).catch(() => false);
    if (!confirmed) return;
    loading.value = true;
    try {
        await revokePostgresqlUser({ database: database.value, name: name.value, username: user.username });
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        emit('search');
    } catch {
        return;
    } finally {
        await load();
    }
};
const remove = async () => {
    if (deleteConfirmInput.value !== deleteUser.username) return;
    deleting.value = true;
    try {
        await deletePostgresqlUser({ database: database.value, username: deleteUser.username });
        MsgSuccess(i18n.global.t('commons.msg.deleteSuccess'));
        deleteDialogVisible.value = false;
        emit('search');
    } catch {
        return;
    } finally {
        deleting.value = false;
        await load();
    }
};
const userButtons = [
    {
        label: i18n.global.t('commons.button.edit'),
        permission: true,
        show: (row: Database.PostgresqlUser) => !row.protected,
        click: (row: Database.PostgresqlUser) => openEdit(row),
    },
    {
        label: i18n.global.t('database.revokeAuthorization'),
        permission: true,
        show: (row: Database.PostgresqlUser) => !!name.value && !row.superUser && !row.protected,
        click: (row: Database.PostgresqlUser) => revoke(row),
    },
    {
        label: i18n.global.t('commons.button.delete'),
        permission: true,
        show: (row: Database.PostgresqlUser) => !name.value && !row.protected,
        click: (row: Database.PostgresqlUser) => openDeleteUserDialog(row),
    },
];
defineExpose({ acceptParams });
</script>

<style lang="scss" scoped>
.authorization-hint {
    margin-bottom: 16px;
}
.user-name-row {
    display: flex;
    align-items: center;
    gap: 6px;
    min-width: 0;
}
.user-name-text {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
.superuser-tag {
    flex-shrink: 0;
}
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
    overflow-wrap: anywhere;
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
.postgresql-bind-user-select-popper {
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
