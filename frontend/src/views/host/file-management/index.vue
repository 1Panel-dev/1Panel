<template>
    <div
        class="file-management-page h-full"
        :class="{ 'is-drag-over': isDragOver }"
        ref="fileTableRef"
        @dragover="handleDragover"
        @drop="handleDrop"
        @dragleave="handleDragleave"
    >
        <el-tabs
            type="card"
            class="file-tabs"
            v-model="editableTabsKey"
            @tab-change="changeTab"
            @tab-remove="removeTab"
        >
            <el-tab-pane
                closable
                v-for="item in editableTabs"
                :key="item.id"
                :label="item.name == '' ? $t('file.root') : item.name"
                :name="item.id"
            >
                <div class="flex flex-wrap items-center gap-3 pb-3 pt-0.5" ref="toolRef">
                    <div class="flex shrink-0 items-center gap-1.5 file-navigation__actions">
                        <el-tooltip :content="$t('file.back')" placement="top">
                            <el-button icon="Back" @click="back" circle />
                        </el-tooltip>
                        <el-tooltip :content="$t('file.right')" placement="top">
                            <el-button icon="Right" @click="right" circle />
                        </el-tooltip>
                        <el-tooltip :content="$t('file.top')" placement="top">
                            <el-button icon="Top" @click="top" circle :disabled="paths.length == 0" />
                        </el-tooltip>
                        <el-tooltip :content="$t('commons.button.refresh')" placement="top">
                            <el-button icon="Refresh" circle @click="search" />
                        </el-tooltip>
                        <el-tooltip
                            :content="req.showHidden ? $t('file.noShowHide') : $t('file.showHide')"
                            placement="top"
                        >
                            <el-button
                                class="btn"
                                circle
                                :type="req.showHidden ? '' : 'primary'"
                                :icon="req.showHidden ? View : Hide"
                                @click="viewHideFile"
                            />
                        </el-tooltip>
                    </div>
                    <div class="min-w-0 flex-1 hidden sm:block" :ref="(el) => setPathRef(item.id, el)">
                        <div v-show="!searchableStatus" @click="searchableStatus = true" class="address-bar">
                            <div ref="breadCrumbRef" class="flex items-center address-url">
                                <span class="breadcrumb-root">
                                    <el-link @click.stop="jump('/')">
                                        <el-icon :size="20"><HomeFilled /></el-icon>
                                    </el-link>
                                </span>
                                <span
                                    v-for="(path, index) in breadcrumbVisiblePaths"
                                    :key="path.url"
                                    class="breadcrumb-item inline-flex min-w-0 items-center"
                                >
                                    <span class="arrow">></span>
                                    <template v-if="index === 1 && breadcrumbHiddenPaths.length > 0">
                                        <el-dropdown>
                                            <span
                                                class="path-segment path-segment--overflow cursor-pointer pathname focus:outline-none focus-visible:outline-none"
                                                @click.stop
                                            >
                                                ..
                                            </span>
                                            <template #dropdown>
                                                <el-dropdown-menu>
                                                    <el-dropdown-item
                                                        v-for="hidePath in breadcrumbHiddenPaths"
                                                        :key="hidePath.url"
                                                        @click.stop="jump(hidePath.url)"
                                                    >
                                                        <el-tooltip
                                                            class="box-item"
                                                            effect="dark"
                                                            :content="hidePath.name"
                                                            placement="bottom"
                                                        >
                                                            {{
                                                                hidePath.name.length > 25
                                                                    ? hidePath.name.substring(0, 22) + '...'
                                                                    : hidePath.name
                                                            }}
                                                        </el-tooltip>
                                                    </el-dropdown-item>
                                                </el-dropdown-menu>
                                            </template>
                                        </el-dropdown>
                                        <span class="arrow">></span>
                                        <el-tooltip
                                            class="box-item"
                                            effect="dark"
                                            :content="path.name"
                                            placement="bottom"
                                        >
                                            <el-link
                                                class="path-segment cursor-pointer pathname"
                                                @click.stop="jump(path.url)"
                                            >
                                                {{ formatBreadcrumbName(path, index) }}
                                            </el-link>
                                        </el-tooltip>
                                    </template>
                                    <template v-else>
                                        <el-tooltip
                                            class="box-item"
                                            effect="dark"
                                            :content="path.name"
                                            placement="bottom"
                                        >
                                            <el-link
                                                class="path-segment cursor-pointer pathname"
                                                @click.stop="jump(path.url)"
                                            >
                                                {{ formatBreadcrumbName(path, index) }}
                                            </el-link>
                                        </el-tooltip>
                                    </template>
                                </span>
                            </div>
                        </div>
                        <el-input
                            :ref="(el) => setSearchableInputRef(item.id, el)"
                            v-show="searchableStatus"
                            v-model="searchablePath"
                            @blur="searchableInputBlur"
                            class="address-input"
                            @keyup.enter="
                                jump(searchablePath);
                                searchableStatus = false;
                            "
                        />
                    </div>
                    <div class="min-w-0 flex-1 sm:hidden block">
                        <div class="address-bar">
                            <div class="flex items-center address-url">
                                <span class="breadcrumb-root">
                                    <el-link @click.stop="jump('/')">
                                        <el-icon :size="20"><HomeFilled /></el-icon>
                                    </el-link>
                                </span>
                                <span
                                    v-for="(path, index) in breadcrumbVisiblePaths"
                                    :key="path.url"
                                    class="breadcrumb-item inline-flex min-w-0 items-center"
                                >
                                    <span class="arrow">></span>
                                    <template v-if="index === 1 && breadcrumbHiddenPaths.length > 0">
                                        <el-dropdown>
                                            <span
                                                class="path-segment path-segment--overflow cursor-pointer pathname focus:outline-none focus-visible:outline-none"
                                                @click.stop
                                            >
                                                ..
                                            </span>
                                            <template #dropdown>
                                                <el-dropdown-menu>
                                                    <el-dropdown-item
                                                        v-for="hidePath in breadcrumbHiddenPaths"
                                                        :key="hidePath.url"
                                                        @click.stop="jump(hidePath.url)"
                                                    >
                                                        <el-tooltip
                                                            class="box-item"
                                                            effect="dark"
                                                            :content="hidePath.name"
                                                            placement="bottom"
                                                        >
                                                            {{
                                                                hidePath.name.length > 25
                                                                    ? hidePath.name.substring(0, 22) + '...'
                                                                    : hidePath.name
                                                            }}
                                                        </el-tooltip>
                                                    </el-dropdown-item>
                                                </el-dropdown-menu>
                                            </template>
                                        </el-dropdown>
                                    </template>
                                    <template v-else>
                                        <span class="arrow">></span>
                                        <el-tooltip
                                            class="box-item"
                                            effect="dark"
                                            :content="path.name"
                                            placement="bottom"
                                        >
                                            <el-link
                                                class="path-segment cursor-pointer pathname"
                                                @click.stop="jump(path.url)"
                                            >
                                                {{ formatBreadcrumbName(path, index) }}
                                            </el-link>
                                        </el-tooltip>
                                    </template>
                                </span>
                            </div>
                        </div>
                    </div>
                    <div class="flex w-full flex-wrap items-center justify-start gap-2 xl:w-auto xl:flex-nowrap">
                        <div class="w-full min-w-0 sm:w-[300px]">
                            <el-input
                                v-model="req.search"
                                clearable
                                @clear="search()"
                                @keydown.enter="search()"
                                :placeholder="$t('file.search')"
                            >
                                <template #prepend>
                                    <el-checkbox v-model="req.containSub">
                                        {{ $t('file.sub') }}
                                    </el-checkbox>
                                </template>
                                <template #append>
                                    <el-button icon="Search" @click="search" round />
                                </template>
                            </el-input>
                        </div>
                        <el-button class="max-w-20" plain type="primary" @click="openAiSearchDrawer">
                            {{ $t('file.aiSearch') }}
                        </el-button>
                    </div>
                </div>
                <LayoutContent class="file-layout" :title="$t('menu.files')" v-loading="loading">
                    <template #prompt>
                        <el-alert type="info" :closable="false">
                            <template #title>
                                <span class="input-help whitespace-break-spaces">
                                    {{ $t('file.fileHelper') }}
                                </span>
                            </template>
                        </el-alert>
                    </template>
                    <template #leftToolBar>
                        <div class="flex max-w-full flex-wrap items-center gap-2">
                            <el-dropdown @command="handleCreate">
                                <el-button type="primary">
                                    {{ $t('commons.button.create') }}
                                    <el-icon><arrow-down /></el-icon>
                                </el-button>
                                <template #dropdown>
                                    <el-dropdown-menu>
                                        <el-dropdown-item command="dir">
                                            <svg-icon iconName="p-file-folder"></svg-icon>
                                            {{ $t('file.dir') }}
                                        </el-dropdown-item>
                                        <el-dropdown-item command="file">
                                            <svg-icon iconName="p-file-normal"></svg-icon>
                                            {{ $t('menu.files') }}
                                        </el-dropdown-item>
                                    </el-dropdown-menu>
                                </template>
                            </el-dropdown>
                            <el-dropdown>
                                <el-button>
                                    {{ $t('commons.button.upload') }}/{{ $t('commons.button.download') }}
                                    <el-icon><arrow-down /></el-icon>
                                </el-button>
                                <template #dropdown>
                                    <el-dropdown-menu>
                                        <el-dropdown-item @click="openUpload">
                                            <el-icon><ElUpload /></el-icon>
                                            {{ $t('commons.button.upload') }}
                                        </el-dropdown-item>
                                        <el-dropdown-item @click="openWget">
                                            <el-icon><ElDownload /></el-icon>
                                            {{ $t('file.remoteFile') }}
                                        </el-dropdown-item>
                                    </el-dropdown-menu>
                                </template>
                            </el-dropdown>
                            <el-button-group class="file-utility-group">
                                <el-button class="btn" @click="openRecycleBin">
                                    {{ $t('file.recycleBin') }}
                                </el-button>
                                <el-button class="btn" @click="toTerminal">
                                    {{ $t('menu.terminal') }}
                                </el-button>
                                <el-popover
                                    placement="bottom"
                                    :width="250"
                                    trigger="hover"
                                    @before-enter="getFavorites"
                                >
                                    <template #reference>
                                        <el-button @click="openFavorite">
                                            {{ $t('file.favorite') }}
                                        </el-button>
                                    </template>
                                    <div class="favorite-item">
                                        <el-table :data="favorites">
                                            <el-table-column prop="name">
                                                <template #default="{ row }">
                                                    <div class="flex justify-between items-center group">
                                                        <el-tooltip
                                                            class="box-item"
                                                            effect="dark"
                                                            :content="row.path"
                                                            placement="left"
                                                        >
                                                            <span
                                                                class="table-link text-ellipsis"
                                                                @click="toFavorite(row)"
                                                                type="primary"
                                                            >
                                                                <svg-icon
                                                                    v-if="row.isDir"
                                                                    className="table-icon"
                                                                    iconName="p-file-folder"
                                                                ></svg-icon>
                                                                <svg-icon
                                                                    v-else
                                                                    className="table-icon"
                                                                    iconName="p-file-normal"
                                                                ></svg-icon>
                                                                {{ row.name }}
                                                            </span>
                                                        </el-tooltip>
                                                        <el-icon
                                                            class="hidden group-hover:block cursor-pointer"
                                                            v-if="!row.isDir"
                                                            @click="jump(row.path)"
                                                        >
                                                            <FolderOpened />
                                                        </el-icon>
                                                    </div>
                                                </template>
                                            </el-table-column>
                                        </el-table>
                                    </div>
                                </el-popover>

                                <el-button class="file-tool">
                                    <el-dropdown>
                                        <template #default>
                                            <el-button
                                                link
                                                class="!w-full !h-full p-2 focus-visible:!outline-none cursor-pointer transition-colors"
                                            >
                                                {{ $t('file.fileTools') }}
                                            </el-button>
                                        </template>
                                        <template #dropdown>
                                            <el-dropdown-menu>
                                                <el-dropdown-item @click="openShareList">
                                                    {{ $t('file.shareList') }}
                                                </el-dropdown-item>
                                                <el-dropdown-item @click="openFileHistoryCenter">
                                                    {{ $t('file.history') }}
                                                </el-dropdown-item>
                                            </el-dropdown-menu>
                                        </template>
                                    </el-dropdown>
                                </el-button>
                                <el-button class="btn" @click="calculateSize(req.path)" :loading="disableBtn">
                                    {{ $t('file.calculate') }}
                                </el-button>
                                <template v-if="hostMount.length == 1">
                                    <el-button class="btn" @click.stop="jump(hostMount[0]?.path)">
                                        {{ hostMount[0]?.path }} ({{ $t('file.root') }})
                                        {{ formatFileSize(hostMount[0]?.free) }}
                                    </el-button>
                                </template>
                                <template v-else>
                                    <el-dropdown>
                                        <el-button class="btn">
                                            {{ hostMount[0]?.path }} ({{ $t('file.root') }})
                                            {{ formatFileSize(hostMount[0]?.free) }}
                                        </el-button>
                                        <template #dropdown>
                                            <el-dropdown-menu>
                                                <template v-for="(mount, index) in hostMount" :key="mount.path">
                                                    <el-dropdown-item v-if="index == 0" @click.stop="jump(mount.path)">
                                                        {{ mount.path }} ({{ $t('file.root') }})
                                                        {{ formatFileSize(mount.free) }}
                                                    </el-dropdown-item>
                                                    <el-dropdown-item v-if="index != 0" @click.stop="jump(mount.path)">
                                                        {{ mount.path }} ({{ $t('home.mount') }})
                                                        {{ formatFileSize(mount.free) }}
                                                    </el-dropdown-item>
                                                </template>
                                            </el-dropdown-menu>
                                        </template>
                                    </el-dropdown>
                                </template>
                            </el-button-group>

                            <el-badge :value="processCount" class="btn" v-if="processCount > 0">
                                <el-button class="btn" @click="openProcess">
                                    {{ $t('file.wgetTask') }}
                                </el-button>
                            </el-badge>
                        </div>
                    </template>
                    <template #rightToolBar>
                        <div
                            :ref="(el) => setBtnWrapperRef(item.id, el)"
                            :class="[
                                'file-batch-toolbar flex max-w-full flex-nowrap items-center gap-2',
                                isRightToolbarWrapped ? 'is-toolbar-wrapped w-full justify-start' : 'justify-end',
                            ]"
                        >
                            <div class="flex min-w-0 flex-nowrap items-center gap-2">
                                <el-button-group class="copy-button" v-if="moveOpen">
                                    <el-tooltip
                                        class="box-item"
                                        effect="dark"
                                        :content="$t('file.paste')"
                                        placement="bottom"
                                    >
                                        <el-button plain @click="openPaste">
                                            {{ $t('file.paste') }}({{ fileMove.count }})
                                        </el-button>
                                    </el-tooltip>
                                    <el-tooltip
                                        class="box-item"
                                        effect="dark"
                                        :content="$t('commons.button.cancel')"
                                        placement="bottom"
                                    >
                                        <el-button plain class="close" icon="Close" @click="closeMove"></el-button>
                                    </el-tooltip>
                                </el-button-group>
                                <template v-if="visibleButtons.length == 0">
                                    <el-dropdown v-if="moreButtons.length">
                                        <el-button>
                                            {{ $t('tabs.more') }}
                                            <el-icon><arrow-down /></el-icon>
                                        </el-button>
                                        <template #dropdown>
                                            <el-dropdown-menu>
                                                <el-dropdown-item
                                                    v-for="btn in moreButtons"
                                                    :key="btn.label"
                                                    @click="btn.action"
                                                    :disabled="selects.length === 0"
                                                >
                                                    {{ $t(btn.label) }}
                                                </el-dropdown-item>
                                            </el-dropdown-menu>
                                        </template>
                                    </el-dropdown>
                                </template>
                                <template v-if="visibleButtons.length > 0">
                                    <el-button-group class="flex max-w-full flex-nowrap items-center">
                                        <template v-for="btn in visibleButtons" :key="btn.label">
                                            <el-button plain @click="btn.action" :disabled="selects.length === 0">
                                                {{ $t(btn.label) }}
                                            </el-button>
                                        </template>
                                        <el-dropdown v-if="moreButtons.length">
                                            <el-button>
                                                {{ $t('tabs.more') }}
                                                <el-icon><arrow-down /></el-icon>
                                            </el-button>
                                            <template #dropdown>
                                                <el-dropdown-menu>
                                                    <el-dropdown-item
                                                        v-for="btn in moreButtons"
                                                        :key="btn.label"
                                                        @click="btn.action"
                                                        :disabled="selects.length === 0"
                                                    >
                                                        {{ $t(btn.label) }}
                                                    </el-dropdown-item>
                                                </el-dropdown-menu>
                                            </template>
                                        </el-dropdown>
                                    </el-button-group>
                                </template>
                                <fu-table-column-select
                                    :columns="columns"
                                    trigger="hover"
                                    :title="$t('commons.table.selectColumn')"
                                    popper-class="popper-class"
                                    :only-icon="true"
                                />
                            </div>
                        </div>
                    </template>
                    <template #main>
                        <ComplexTable
                            class="file-table"
                            :pagination-config="paginationConfig"
                            v-model:selects="selects"
                            :ref="(el) => setTableRef(item.id, el)"
                            :data="data"
                            @search="search"
                            @sort-change="changeSort"
                            @cell-mouse-enter="showFavorite"
                            @cell-mouse-leave="hideFavorite"
                            :heightDiff="heightDiff"
                            :right-buttons="rightButtons"
                            :columns="columns"
                            localKey="fileManagementColumn"
                        >
                            <el-table-column type="selection" width="30" />
                            <el-table-column
                                :label="$t('commons.table.name')"
                                min-width="250"
                                fix
                                show-overflow-tooltip
                                :sortable="'custom'"
                                prop="name"
                                :tooltip-options="{
                                    placement: 'bottom-start',
                                }"
                            >
                                <template #default="{ row }">
                                    <div class="file-row">
                                        <div class="file-row__icon">
                                            <svg-icon
                                                v-if="row.isDir"
                                                className="table-icon"
                                                iconName="p-file-folder"
                                            ></svg-icon>
                                            <svg-icon
                                                v-else
                                                className="table-icon"
                                                :iconName="getIconName(row.name, row.extension)"
                                            ></svg-icon>
                                        </div>
                                        <div class="file-name">
                                            <el-input
                                                v-if="fileRename.oldName === row.name && isEdit"
                                                v-model.trim="fileRename.newName"
                                                :ref="(el) => setRenameRef(item.id, el)"
                                                :autofocus="isEdit"
                                                class="table-link table-input"
                                                placeholder="file name"
                                                @keydown.enter="handleRename(row)"
                                                @blur="onRenameBlur($event, row)"
                                            />
                                            <span v-else class="table-link" @click="open(row)" type="primary">
                                                {{ row.name }}
                                            </span>
                                            <span v-if="row.isSymlink">-> {{ row.linkPath }}</span>
                                        </div>
                                        <div class="file-row__actions">
                                            <el-button
                                                v-if="row.shareCode"
                                                link
                                                type="primary"
                                                size="large"
                                                icon="Share"
                                                @click="openShareFile(row)"
                                            ></el-button>
                                        </div>
                                        <div class="file-row__actions">
                                            <el-button
                                                v-if="row.favoriteID > 0"
                                                link
                                                type="warning"
                                                size="large"
                                                icon="StarFilled"
                                                @click="remove(row.favoriteID)"
                                            ></el-button>
                                            <div v-else>
                                                <el-button
                                                    v-if="hoveredRowPath === row.path"
                                                    link
                                                    icon="Star"
                                                    @click="addToFavorite(row)"
                                                ></el-button>
                                            </div>
                                        </div>
                                    </div>
                                </template>
                            </el-table-column>
                            <el-table-column :label="$t('file.mode')" prop="mode" width="80">
                                <template #default="{ row }">
                                    <el-link underline="never" @click="openMode(row)">{{ row.mode }}</el-link>
                                </template>
                            </el-table-column>
                            <el-table-column
                                :label="`${$t('commons.table.user')} / ${$t('file.group')}`"
                                prop="user"
                                show-overflow-tooltip
                                width="200"
                            >
                                <template #default="{ row }">
                                    <el-link underline="never" @click="openChown(row)">
                                        {{ row.user ? row.user : '-' }} ({{ row.uid }}) /
                                        {{ row.group ? row.group : '-' }} ({{ row.gid }})
                                    </el-link>
                                </template>
                            </el-table-column>
                            <el-table-column :label="$t('file.size')" prop="size" width="120" :sortable="'custom'">
                                <template #default="{ row }">
                                    <el-button
                                        type="primary"
                                        link
                                        small
                                        :loading="row.btnLoading"
                                        @click="row.isDir ? getDirSize(row.path) : getFileSize(row.path)"
                                    >
                                        <span v-if="row.isDir">
                                            <span v-if="row.dirSize === undefined">
                                                {{ $t('file.calculate') }}
                                            </span>
                                            <span v-else>{{ formatFileSize(row.dirSize) }}</span>
                                        </span>
                                        <span v-else>
                                            {{ formatFileSize(row.size) }}
                                        </span>
                                    </el-button>
                                </template>
                            </el-table-column>
                            <el-table-column
                                :label="$t('file.updateTime')"
                                prop="modTime"
                                width="180"
                                :formatter="dateFormat"
                                show-overflow-tooltip
                                :sortable="'custom'"
                            ></el-table-column>
                            <el-table-column :label="$t('file.remark')" prop="remark" width="180" show-overflow-tooltip>
                                <template #default="{ row }">
                                    <span>{{ row.remark ? row.remark : '-' }}</span>
                                </template>
                            </el-table-column>
                            <fu-table-operations
                                :max-height="dropdownMaxHeight"
                                :ellipsis="mobile ? 0 : 2"
                                :buttons="tableMoreButtons"
                                :label="$t('commons.table.operate')"
                                :min-width="mobile ? 'auto' : 200"
                                :fixed="mobile ? false : 'right'"
                                width="200"
                                fix
                            />
                            <template #paginationLeft>
                                <div class="file-pagination-summary">
                                    <el-text small>
                                        {{ $t('file.fileDirNum', [dirNum, fileNum]) }}
                                    </el-text>
                                    <el-text small>
                                        {{ $t('file.currentDir') + $t('file.size') + ' ' }}
                                    </el-text>
                                    <el-button type="primary" link small :loading="calculateBtn">
                                        <span v-if="dirTotalSize == -1" @click="getDirTotalSize(req.path)">
                                            {{ $t('file.calculate') }}
                                        </span>
                                        <span v-else>
                                            {{ formatFileSize(dirTotalSize) }}
                                        </span>
                                    </el-button>
                                </div>
                            </template>
                        </ComplexTable>
                    </template>
                </LayoutContent>
            </el-tab-pane>
            <el-tab-pane :name="editableTabsKey" :closable="false" :disabled="editableTabs.length > 6">
                <template #label>
                    <el-icon @click="addTab()"><Plus /></el-icon>
                </template>
            </el-tab-pane>
        </el-tabs>

        <CreateFile ref="createRef" @close="search" />
        <ChangeRole ref="roleRef" @close="search" />
        <Compress ref="compressRef" @close="search" @task-change="handleFileTaskChange('compress', $event)" />
        <Decompress ref="deCompressRef" @close="search" @task-change="handleFileTaskChange('decompress', $event)" />
        <CodeEditor ref="codeEditorRef" @close="search" />
        <FileRename ref="renameRef" @close="search" />
        <Upload ref="uploadRef" @close="search" />
        <Wget ref="wgetRef" @close="closeWget" />
        <Move ref="moveRef" @close="closeMovePage" @loading="onLoading" />
        <Download ref="downloadRef" @close="search" />
        <Process ref="processRef" @close="closeProcess" />
        <Owner ref="chownRef" @close="search"></Owner>
        <Detail ref="detailRef" />
        <DeleteFile ref="deleteRef" @close="search" />
        <RecycleBin ref="recycleBinRef" @close="search" />
        <Favorite ref="favoriteRef" @close="search" @jump="jump" @to-favorite="toFavorite" />
        <ShareList ref="shareListRef" @close="search" @detail="openShareDetail" />
        <FileHistoryDrawer ref="historyDrawerRef" @restored="search" />
        <BatchRole ref="batchRoleRef" @close="search" />
        <VscodeOpenDialog ref="dialogVscodeOpenRef" />
        <Preview ref="previewRef" />
        <TextPreview ref="textPreviewRef" />
        <TerminalDialog ref="dialogTerminalRef" />
        <Convert ref="convertRef" @close="search" />

        <FileAiSearchDrawer
            ref="aiSearchDrawerRef"
            v-model="aiSearchDrawerVisible"
            :list-path="req.path"
            @pick-directory="openAiSearchPathPicker"
            @open-editor="onAiSearchOpenEditor"
        />
        <FileList ref="fileRef" @choose="getSearchPath" />
        <FileShare ref="fileShareRef" @close="search" />
    </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref } from 'vue';
import {
    addFavorite,
    batchGetFileRemarks,
    computeDepthDirSize,
    computeDirSize,
    fileWgetKeys,
    getFileContent,
    checkFile,
    removeFileShare,
    getFilesList,
    setFileRemark,
    removeFavorite,
    renameRile,
    searchFavorite,
    searchHostMount,
} from '@/api/modules/files';
import { computeSize } from '@/utils/size';
import { copyText } from '@/utils/clipboard';
import { dateFormat } from '@/utils/date';
import { downloadFile, getFileType, getIcon, isConvertible } from '@/utils/file';
import { getRandomStr } from '@/utils/id';
import { File } from '@/api/interface/file';
import { Mimetypes } from '@/global/mimetype';
import { resolveEditorLanguage } from '@/utils/file';
import { useRouter } from 'vue-router';
import { MsgSuccess, MsgWarning } from '@/utils/message';
import { useMultipleSearchable } from './hooks/searchable';
import { ResultData } from '@/api/interface';
import { GlobalStore } from '@/store';
import { Download as ElDownload, Upload as ElUpload, View, Hide } from '@element-plus/icons-vue';

import i18n from '@/lang';
import CreateFile from './create/index.vue';
import ChangeRole from './change-role/index.vue';
import Compress from './compress/index.vue';
import Decompress from './decompress/index.vue';
import Upload from './upload/index.vue';
import FileRename from './rename/index.vue';
import CodeEditor from './code-editor/index.vue';
import Wget from './wget/index.vue';
import Move from './move/index.vue';
import Download from './download/index.vue';
import Owner from './chown/index.vue';
import DeleteFile from './delete/index.vue';
import Process from './process/index.vue';
import Detail from './detail/index.vue';
import RecycleBin from './recycle-bin/index.vue';
import Favorite from './favorite/index.vue';
import ShareList from './share-list/index.vue';
import FileHistoryDrawer from './code-editor/history/index.vue';
import BatchRole from './batch-role/index.vue';
import Preview from './preview/index.vue';
import TextPreview from './text-preview/index.vue';
import VscodeOpenDialog from '@/components/vscode-open/index.vue';
import Convert from './convert/index.vue';
import FileAiSearchDrawer from './ai-search/file-ai-search-drawer.vue';
import FileShare from './share/index.vue';
import { debounce } from 'lodash-es';
import TerminalDialog from './terminal/index.vue';
import { Dashboard } from '@/api/interface/dashboard';
import { CompressExtension, MimetypeByExtensionObject } from '@/enums/files';
import type { TabPaneName } from 'element-plus';
import { getComponentInfo } from '@/api/modules/host';
import { routerToNameWithQuery } from '@/utils/router';
import { loadBaseDir } from '@/api/modules/setting';
import FileList from '@/components/file-list/index.vue';

const globalStore = GlobalStore();

interface FilePaths {
    url: string;
    name: string;
}

const fileRef = ref();
const router = useRouter();
const data = ref();
const tableRefs = ref<Record<string, any>>({});
const heightDiff = ref(365);
const fileTableRef = ref<HTMLElement | null>(null);
const dropdownMaxHeight = ref(450);
const baseDir = ref();
const remarkRequestId = ref(0);
const remarkLoadTimer = ref<number | null>(null);
const editableTabsKey = ref('');
const editableTabs = ref([
    { id: '1', name: getLastPath(baseDir.value), path: baseDir.value },
    { id: '2', name: 'home', path: '/home' },
]);

const setTableRef = (key: string, el: any) => {
    if (el) {
        tableRefs.value[key] = el;
    }
};
const getCurrentTable = () => tableRefs.value[editableTabsKey.value];

let selects = ref<any>([]);

const initData = () => ({
    path: '/',
    expand: true,
    showHidden: localStorage.getItem('show-hidden') === 'true',
    page: 1,
    pageSize: 100,
    search: '',
    containSub: false,
    sortBy: 'name',
    sortOrder: 'ascending',
});
let req = reactive(initData());
let loading = ref(false);
const paths = ref<FilePaths[]>([]);
const breadcrumbVisiblePaths = ref<FilePaths[]>([]);
const breadcrumbHiddenPaths = ref<FilePaths[]>([]);
let pathWidth = ref(0);
const history: string[] = [];
let pointer = -1;

const fileCreate = reactive({ path: '/', isDir: false, mode: 0o755 });
const fileCompress = reactive({ files: [''], name: '', dst: '', operate: 'compress' });
const fileDeCompress = reactive({ path: '', name: '', dst: '', type: '' });
const fileEdit = reactive<{
    content: string;
    path: string;
    name: string;
    language: string;
    extension: string;
    initialLine?: number;
}>({ content: '', path: '', name: '', language: 'plaintext', extension: '' });
const filePreview = reactive({ path: '', name: '', extension: '', fileType: '', imageFiles: [], currentNode: '' });
const codeReq = reactive({ path: '', expand: false, page: 1, pageSize: 100, isDetail: false });
const fileUpload = reactive({ path: '' });
const fileRename = reactive({ path: '', oldName: '', newName: '' });
const fileWget = reactive({ path: '' });
const fileMove = reactive({ oldPaths: [''], allNames: [''], type: '', path: '', name: '', count: 0, isDir: false });
const fileConvert = reactive<{
    outputPath: string;
    files: File.ConvertFile[];
}>({
    outputPath: '',
    files: [
        {
            type: '',
            inputFile: '',
            extension: '',
            path: '',
            outputFormat: '',
        },
    ],
});
const ffmpegExist = ref(false);

const aiSearchDrawerVisible = ref(false);
const aiSearchDrawerRef = ref<InstanceType<typeof FileAiSearchDrawer> | null>(null);

const openAiSearchDrawer = () => {
    aiSearchDrawerVisible.value = true;
};

const getSearchPath = (path: string | string[]) => {
    aiSearchDrawerRef.value?.applyPathFromPicker(path);
};

const openAiSearchPathPicker = (path?: string) => {
    fileRef.value.acceptParams({ path: path || req.path, dir: true, multiple: false });
};

const createRef = ref();
const roleRef = ref();
const detailRef = ref();
const compressRef = ref();
const deCompressRef = ref();
const codeEditorRef = ref();
const renameRef = ref();
const uploadRef = ref();
const wgetRef = ref();
const moveRef = ref();
const downloadRef = ref();
const toolRef = ref();
const breadCrumbRef = ref();
const chownRef = ref();
const moveOpen = ref(false);
const deleteRef = ref();
const recycleBinRef = ref();
const favoriteRef = ref();
const shareListRef = ref();
const historyDrawerRef = ref<InstanceType<typeof FileHistoryDrawer> | null>(null);
const hoveredRowPath = ref(null);
const favorites = ref([]);
const batchRoleRef = ref();
const dialogVscodeOpenRef = ref();
const previewRef = ref();
const textPreviewRef = ref();
const processRef = ref();

const MAX_OPEN_SIZE = 10 * 1024 * 1024;
const MAX_DIR_SIZE_CONCURRENT = 2;
const hostMount = ref<Dashboard.DiskInfo[]>([]);
let resizeObserver: ResizeObserver;
let depthSizeToken = 0;
const dirTotalSize = ref(-1);
const disableBtn = ref(false);
const calculateBtn = ref(false);
const dirNum = ref(0);
const fileNum = ref(0);
const imageFiles = ref([]);
const isEdit = ref(false);
const isDragOver = ref(false);
const convertRef = ref();
const fileTaskStatus = reactive<Record<string, string>>({});

const renameRefs = ref<Record<string, any>>({});

const setRenameRef = (key: string, el: any) => {
    if (el) {
        renameRefs.value[key] = el;
    }
};
const getCurrentRename = () => renameRefs.value[editableTabsKey.value];

const pathRefs = ref<Record<string, any>>({});
const columns = ref([]);

const setPathRef = (key: string, el: any) => {
    if (el) {
        pathRefs.value[key] = el;
    }
};
const getCurrentPath = () => pathRefs.value[editableTabsKey.value];

const { searchableStatus, searchablePath, setSearchableInputRef, searchableInputBlur } = useMultipleSearchable(paths);

const paginationConfig = reactive({
    cacheSizeKey: 'file-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('file-page-size')) || 100,
    total: 0,
});

const mobile = computed(() => {
    return globalStore.isMobile();
});

const btnWrapperRefs = ref<Record<string, any>>({});

const setBtnWrapperRef = (key: string, el: any) => {
    if (el) {
        btnWrapperRefs.value[key] = el;
    }
};
const getCurrentBtnWrapper = () => btnWrapperRefs.value[editableTabsKey.value];

const toolButtons = ref([
    {
        label: 'commons.button.copy',
        action: () => openMove('copy'),
    },
    {
        label: 'file.move',
        action: () => openMove('cut'),
    },
    {
        label: 'file.compress',
        action: () => openCompress(selects.value),
    },
    {
        label: 'file.role',
        action: () => openBatchRole(selects.value),
    },
    {
        label: 'commons.button.delete',
        action: () => batchDelFiles(),
    },
]);

const visibleButtons = ref([...toolButtons.value]);
const moreButtons = ref([]);
const isRightToolbarWrapped = ref(false);
const batchButtonMinWidths = [64, 64, 64, 64, 64];
const moreButtonWidth = 76;
const toolbarGap = 8;

const updateButtons = async () => {
    await nextTick();
    const wrapper = getCurrentBtnWrapper();
    if (!wrapper) {
        return;
    }
    const slotWrapper = wrapper.parentElement as HTMLElement | null;
    const titleRow = slotWrapper?.parentElement as HTMLElement | null;
    const rowWidth = titleRow?.clientWidth || slotWrapper?.clientWidth || wrapper.offsetWidth || 0;
    if (!rowWidth) {
        visibleButtons.value = [...toolButtons.value];
        moreButtons.value = [];
        return;
    }

    const pasteEl = wrapper.querySelector<HTMLElement>('.copy-button');
    const leftSibling = slotWrapper?.previousElementSibling as HTMLElement | null;
    const pasteWidth = pasteEl?.offsetWidth || 0;
    const pasteReserve = moveOpen.value ? pasteWidth + toolbarGap : 0;
    const columnSelectWidth = 48;
    const minBatchWidth = moreButtonWidth;
    const sameLineReserve = leftSibling ? leftSibling.offsetWidth + 16 : 0;
    const sameLineAvailable = rowWidth - sameLineReserve - pasteReserve - columnSelectWidth - toolbarGap;
    isRightToolbarWrapped.value = !!leftSibling && sameLineAvailable < minBatchWidth;
    const leftReserve = isRightToolbarWrapped.value ? 0 : sameLineReserve;
    const rightLineWidth = Math.max(0, rowWidth - leftReserve - columnSelectWidth - toolbarGap);
    const availableWidth = Math.max(0, rightLineWidth - pasteReserve);
    let usedWidth = 0;
    let visibleCount = 0;

    for (const buttonWidth of batchButtonMinWidths) {
        const nextUsedWidth = usedWidth + buttonWidth;
        const remainingButtonCount = toolButtons.value.length - visibleCount - 1;
        const reserveMoreWidth = remainingButtonCount > 0 ? moreButtonWidth : 0;
        if (nextUsedWidth + reserveMoreWidth <= availableWidth) {
            usedWidth = nextUsedWidth;
            visibleCount++;
        } else {
            break;
        }
    }

    visibleButtons.value = toolButtons.value.slice(0, visibleCount);
    moreButtons.value = toolButtons.value.slice(visibleCount);
};

const search = async () => {
    dirTotalSize.value = -1;
    await getWgetProcess();
    loading.value = true;
    if (req.search != '') {
        req.sortBy = 'name';
        req.sortOrder = 'ascending';
        getCurrentTable().clearSort();
    }
    req.page = paginationConfig.currentPage;
    req.pageSize = paginationConfig.pageSize;
    await getFilesList(req)
        .then((res) => {
            handleSearchResult(res);
        })
        .finally(() => {
            loading.value = false;
        });
};

const searchFile = async () => {
    loading.value = true;
    dirTotalSize.value = -1;
    try {
        return await getFilesList(req);
    } finally {
        loading.value = false;
    }
};

const handleSearchResult = (res: ResultData<File.File>) => {
    data.value = res.data.items || [];
    paginationConfig.total = res.data.itemTotal;
    dirNum.value = data.value.filter((item) => item.isDir).length;
    fileNum.value = data.value.filter((item) => !item.isDir).length;
    if (res.data.path) {
        req.path = res.data.path;
    }
    scheduleRemarkLoad();
};

const normalizeFilePath = (filePath: string) => {
    if (!filePath) {
        return '/';
    }
    const normalized = `/${filePath.split('/').filter(Boolean).join('/')}`;
    return normalized === '' ? '/' : normalized;
};

const findExistingPath = async (filePath: string) => {
    const segments = normalizeFilePath(filePath).split('/').filter(Boolean);
    while (segments.length > 0) {
        const current = `/${segments.join('/')}`;
        try {
            const res = await checkFile(current, false);
            if (res.data) {
                return current;
            }
        } catch {
            // ignore check errors
        }
        segments.pop();
    }
    return '/';
};

const loadInitialExistingPath = async (url: string) => {
    const existingPath = await findExistingPath(url);
    if (existingPath === normalizeFilePath(url)) {
        return;
    }
    const { pageSize: oldPageSize, sortBy: oldSortBy, sortOrder: oldSortOrder, showHidden } = req;
    Object.assign(req, initData(), {
        path: existingPath,
        containSub: false,
        search: '',
        pageSize: oldPageSize,
        sortBy: oldSortBy,
        sortOrder: oldSortOrder,
        showHidden,
    });
    globalStore.lastFilePath = req.path;
    getPaths(req.path);
    updateTab(req.path);
    paths.value = buildPaths(req.path);
    resetPaths();
    MsgWarning(i18n.global.t('commons.res.notFound'));
};

const viewHideFile = async () => {
    req.showHidden = !req.showHidden;
    localStorage.setItem('show-hidden', req.showHidden ? 'true' : 'false');
    let searchResult = await searchFile();
    handleSearchResult(searchResult);
};

const open = async (row: File.File) => {
    hideRightMenu();
    calculateBtn.value = false;
    disableBtn.value = false;
    if (row.isDir) {
        if (row.name.indexOf('.1panel_clash') > -1) {
            MsgWarning(i18n.global.t('file.clashOpenAlert'));
            return;
        }
        const name = row.name;
        if (req.path.endsWith('/')) {
            req.path = req.path + name;
        } else {
            req.path = req.path + '/' + name;
        }
        paths.value.push({
            url: req.path,
            name: name,
        });
        await jump(req.path);
    } else {
        openView(row);
    }
};

const copyDir = (row: File.File) => {
    if (row?.path) {
        copyText(row?.path);
    }
};

const formatBreadcrumbName = (path: FilePaths, index: number) => {
    const isLast = index === breadcrumbVisiblePaths.value.length - 1;
    if (isLast || path.name.length <= 25) {
        return path.name;
    }
    return `${path.name.substring(0, 22)}...`;
};

const getBreadcrumbElement = () => {
    return getCurrentPath()?.querySelector<HTMLElement>('.address-url') || breadCrumbRef.value;
};

const resetPaths = () => {
    breadcrumbVisiblePaths.value = [...paths.value];
    breadcrumbHiddenPaths.value = [];
};

const handlePath = async () => {
    resetPaths();
    if (paths.value.length <= 2) {
        return;
    }

    await nextTick();
    const breadcrumbEl = getBreadcrumbElement();
    const pathEl = getCurrentPath();
    const maxWidth = (pathEl?.clientWidth || toolRef.value?.clientWidth || 0) - 24;
    if (!breadcrumbEl || maxWidth <= 0) {
        return;
    }

    let hiddenCount = 0;
    const maxHiddenCount = Math.max(paths.value.length - 2, 0);
    while (hiddenCount < maxHiddenCount && breadcrumbEl.scrollWidth > maxWidth) {
        hiddenCount++;
        breadcrumbHiddenPaths.value = paths.value.slice(1, 1 + hiddenCount);
        breadcrumbVisiblePaths.value = [paths.value[0], ...paths.value.slice(1 + hiddenCount)];
        await nextTick();
    }
};

const resizeHandler = debounce(() => {
    handlePath();
}, 100);

const btnResizeHandler = debounce(() => {
    updateButtons();
}, 100);

const observeResize = () => {
    const el = getCurrentPath();
    const ele = getCurrentBtnWrapper();
    const titleRow = ele?.parentElement?.parentElement as HTMLElement | null;
    if (!el && !ele && !titleRow) return;

    const observe = new ResizeObserver((entries) => {
        const isElChanged = entries.some((entry) => entry.target === el);
        const isEleChanged = entries.some((entry) => entry.target === ele);
        const isTitleChanged = entries.some((entry) => entry.target === titleRow);

        if (isElChanged) resizeHandler();
        if (isEleChanged || isTitleChanged) btnResizeHandler();
        updateHeight();
    });

    if (el) observe.observe(el);
    if (ele) observe.observe(ele);
    if (titleRow) observe.observe(titleRow);

    resizeObserver = observe;
};

function watchTitleHeight() {
    const el = document.querySelector<HTMLElement>('.content-container__title');
    if (el) {
        let titleHeight = el.offsetHeight < 40 ? 40 : 80;
        heightDiff.value = 325 + titleHeight;
    }
}

const right = () => {
    if (pointer < history.length - 1) {
        pointer++;
        let url = history[pointer];
        backForwardJump(url);
    }
};

const back = () => {
    if (pointer > 0) {
        pointer--;
        let url = history[pointer];
        backForwardJump(url);
    }
};

const top = () => {
    if (paths.value.length > 0) {
        let url = '/';
        if (paths.value.length >= 2) {
            url = paths.value[paths.value.length - 2].url;
        }
        jump(url);
    }
};

const jump = async (url: string) => {
    hideRightMenu();
    history.splice(pointer + 1);
    history.push(url);
    pointer = history.length - 1;

    const { path: oldUrl, pageSize: oldPageSize, sortBy: oldSortBy, sortOrder: oldSortOrder, showHidden } = req;
    Object.assign(req, initData(), {
        path: url,
        containSub: false,
        search: '',
        pageSize: oldPageSize,
        sortBy: oldSortBy,
        sortOrder: oldSortOrder,
        showHidden,
    });
    let searchResult = await searchFile();
    if (!searchResult.data.path) {
        req.path = oldUrl;
        globalStore.lastFilePath = req.path;
        MsgWarning(i18n.global.t('commons.res.notFound'));
        return;
    }
    req.path = searchResult.data.path;
    globalStore.lastFilePath = req.path;
    handleSearchResult(searchResult);
    getPaths(req.path);
    updateTab(req.path);
    await nextTick(function () {
        handlePath();
    });
};

const backForwardJump = async (url: string) => {
    const { pageSize: oldPageSize, sortBy: oldSortBy, sortOrder: oldSortOrder, showHidden } = req;
    Object.assign(req, initData(), {
        path: url,
        containSub: false,
        search: '',
        pageSize: oldPageSize,
        sortBy: oldSortBy,
        sortOrder: oldSortOrder,
        showHidden,
    });
    let searchResult = await searchFile();
    handleSearchResult(searchResult);
    getPaths(req.path);
    updateTab(req.path);
    await nextTick(function () {
        handlePath();
    });
};

const getPaths = (reqPath: string | undefined | null) => {
    const pathString = reqPath || '';
    const pathArray = pathString.split('/').filter((p) => p !== '');

    const breadcrumbs = [];
    let base = '';

    for (const p of pathArray) {
        base = `${base}/${p}`;
        breadcrumbs.push({
            url: base,
            name: p,
        });
    }

    paths.value = breadcrumbs;
    resetPaths();
};

const handleCreate = (command: string) => {
    fileCreate.path = req.path;
    fileCreate.isDir = command === 'dir';
    createRef.value.acceptParams(fileCreate);
};

const delFile = async (row: File.File | null) => {
    deleteRef.value.acceptParams([row]);
};

const batchDelFiles = () => {
    deleteRef.value.acceptParams(selects.value);
};

const formatFileSize = (size: number) => {
    return computeSize(size);
};

const dirSizeQueue: Array<() => Promise<void>> = [];
const dirSizeLoadingPaths = new Set<string>();
let activeDirSizeRequests = 0;

const flushDirSizeQueue = () => {
    while (activeDirSizeRequests < MAX_DIR_SIZE_CONCURRENT && dirSizeQueue.length > 0) {
        const task = dirSizeQueue.shift()!;
        activeDirSizeRequests++;
        task().finally(() => {
            activeDirSizeRequests--;
            flushDirSizeQueue();
        });
    }
};

const enqueueDirSizeTask = (task: () => Promise<void>) => {
    return new Promise<void>((resolve, reject) => {
        dirSizeQueue.push(async () => {
            try {
                await task();
                resolve();
            } catch (err) {
                reject(err);
            }
        });
        flushDirSizeQueue();
    });
};

const getFileSize = async (path: string) => {
    codeReq.path = path;
    codeReq.expand = true;
    codeReq.isDetail = true;
    updateByPath(path, { btnLoading: true });
    try {
        const res = await getFileContent(codeReq);
        updateByPath(path, { dirSize: res.data.size });
    } finally {
        updateByPath(path, { btnLoading: false });
    }
};

const getDirSize = async (path: string) => {
    if (dirSizeLoadingPaths.has(path)) {
        return;
    }
    const req = {
        path: path,
    };
    dirSizeLoadingPaths.add(path);
    updateByPath(path, { btnLoading: true });
    try {
        await enqueueDirSizeTask(async () => {
            const res = await computeDirSize(req);
            updateByPath(path, { dirSize: res.data.size });
        });
    } finally {
        dirSizeLoadingPaths.delete(path);
        updateByPath(path, { btnLoading: false });
    }
};

const updateByPath = (path: string, patch: Partial<(typeof data.value)[0]>) => {
    data.value = data.value.map((item) => (item.path === path ? { ...item, ...patch } : item));
};

const getDirTotalSize = async (path: string) => {
    const sizeReq = {
        path: path,
    };
    calculateBtn.value = true;
    try {
        const res = await computeDirSize(sizeReq);
        dirTotalSize.value = res.data.size;
    } finally {
        calculateBtn.value = false;
    }
};

const calculateSize = (path: string) => {
    const token = ++depthSizeToken;
    const sizeReq = { path };
    disableBtn.value = true;
    setTimeout(async () => {
        try {
            const res = await computeDepthDirSize(sizeReq);
            if (token !== depthSizeToken || req.path !== path) {
                return;
            }
            const sizeMap = new Map(res.data.map((dir) => [dir.path, dir.size]));
            data.value = data.value.map((item) =>
                sizeMap.has(item.path) ? { ...item, dirSize: sizeMap.get(item.path)! } : item,
            );
        } catch (err) {
            console.error('Error computing dir size:', err);
        } finally {
            if (token === depthSizeToken) {
                disableBtn.value = false;
            }
        }
    }, 0);
};

const getIconName = (name: string, extension: string) => {
    return getIcon(getFileExtension(name, extension));
};

const openMode = (item: File.File) => {
    roleRef.value.acceptParams(item);
};

const openChown = (item: File.File) => {
    chownRef.value.acceptParams(item);
};

const openCompress = (items: File.File[]) => {
    const paths = [];
    for (const item of items) {
        paths.push(item.path);
    }
    fileCompress.files = paths;
    if (paths.length === 1) {
        fileCompress.name = items[0].name;
    } else {
        fileCompress.name = getRandomStr(6);
    }
    fileCompress.dst = req.path;

    compressRef.value.acceptParams(fileCompress);
};

const openDeCompress = (item: File.File) => {
    const extension = getFileExtension(item.name, item.extension);
    const mimeType = item.mimeType || MimetypeByExtensionObject[extension];
    const typeByMime = mimeType ? Mimetypes.get(mimeType) : undefined;
    const typeByExtension = getEnumKeyByValue(extension);

    if (typeByMime && (!typeByExtension || CompressExtension[typeByMime] === extension)) {
        fileDeCompress.type = typeByMime;
    } else if (typeByExtension) {
        fileDeCompress.type = typeByExtension;
    } else {
        MsgWarning(i18n.global.t('file.canNotDeCompress'));
        return;
    }

    fileDeCompress.name = item.name;
    fileDeCompress.path = item.path;
    fileDeCompress.dst = req.path;

    deCompressRef.value.acceptParams(fileDeCompress);
};

function getEnumKeyByValue(value: string): keyof typeof CompressExtension | undefined {
    const normalizedValue = value.toLowerCase();
    return (Object.keys(CompressExtension) as Array<keyof typeof CompressExtension>).find(
        (k) => CompressExtension[k] === normalizedValue,
    );
}

const sortedCompressExtensions = Object.values(CompressExtension).sort((a, b) => b.length - a.length);

const getFileExtension = (name: string, extension?: string): string => {
    const lowerName = name?.toLowerCase().split('?')[0] ?? '';
    if (lowerName.startsWith('.') && lowerName.indexOf('.', 1) === -1) {
        return extension.toLowerCase();
    }
    const compoundMatch = sortedCompressExtensions.find((compressExtension) => lowerName.endsWith(compressExtension));
    if (compoundMatch) {
        return compoundMatch;
    }

    if (extension) {
        const lowerExt = extension.toLowerCase();
        return lowerExt.startsWith('.') ? lowerExt : `.${lowerExt}`;
    }

    const extensionIndex = lowerName.lastIndexOf('.');
    return extensionIndex === -1 ? '' : lowerName.slice(extensionIndex);
};

const openView = (item: File.File) => {
    const fileType = getFileType(item.extension);
    if (fileType === 'image') {
        imageFiles.value = data.value
            .filter((item) => !item.isDir)
            .filter((item) => getFileType(item.extension) == 'image')
            .map((item) => (item.isSymlink ? item.linkPath : item.path));
    }

    const previewTypes = ['image', 'video', 'audio', 'word', 'excel'];
    if (previewTypes.includes(fileType)) {
        return openPreview(item, fileType);
    }

    if (fileType === 'compress') {
        return openDeCompress(item);
    }

    const path = item.isSymlink ? item.linkPath : item.path;
    if (item.size > MAX_OPEN_SIZE) {
        return openTextPreview(path, item.name);
    }

    const actionMap = {
        text: () => openCodeEditor(path),
    };

    return actionMap[fileType] ? actionMap[fileType](item) : openCodeEditor(path);
};

const openPreview = (item: File.File, fileType: string) => {
    if (item.mode.toString() == '-' && item.user == '-' && item.group == '-') {
        MsgWarning(i18n.global.t('file.fileCanNotRead'));
        return;
    }
    filePreview.path = item.isSymlink ? item.linkPath : item.path;
    filePreview.name = item.name;
    filePreview.extension = item.extension;
    filePreview.fileType = fileType;
    filePreview.imageFiles = imageFiles.value;
    filePreview.currentNode = globalStore.currentNode;

    previewRef.value.acceptParams(filePreview);
};

const openPathInCodeEditor = (
    path: string,
    opts?: {
        initialLine?: number;
    },
) => {
    if (!path) {
        return;
    }
    codeReq.path = path;
    codeReq.expand = true;

    const line = opts?.initialLine && opts.initialLine > 0 ? Math.floor(opts.initialLine) : undefined;

    getFileContent(codeReq)
        .then((res) => {
            fileEdit.content = res.data.content;
            fileEdit.path = res.data.path;
            fileEdit.name = res.data.name;
            fileEdit.extension = res.data.extension;
            fileEdit.language = resolveEditorLanguage(res.data.path, res.data.extension, res.data.name);
            fileEdit.initialLine = line;
            codeEditorRef.value.acceptParams(fileEdit);
            fileEdit.initialLine = undefined;
        })
        .catch(() => {});
};

const onAiSearchOpenEditor = (payload: { path: string; initialLine?: number }) => {
    openPathInCodeEditor(payload.path, { initialLine: payload.initialLine });
};

const openCodeEditor = (path: string) => {
    openPathInCodeEditor(path);
};

const openTextPreview = (path: string, name: string) => {
    textPreviewRef.value.acceptParams({ path, name });
};

const openUpload = () => {
    fileUpload.path = req.path;
    uploadRef.value.acceptParams(fileUpload);
};

const openWget = () => {
    fileWget.path = req.path;
    wgetRef.value.acceptParams(fileWget);
};

const openBatchRole = (items: File.File[]) => {
    batchRoleRef.value.acceptParams({ files: items });
};

const closeWget = (submit: boolean) => {
    search();
    if (submit) {
        openProcess();
    }
};

const closeMovePage = (submit: boolean) => {
    if (submit) {
        search();
        closeMove();
    }
};

const openProcess = () => {
    processRef.value.acceptParams();
};

const closeProcess = () => {
    search();
    getWgetProcess();
    setTimeout(() => {
        getWgetProcess();
    }, 1000);
};

const processCount = ref(0);
const getWgetProcess = async () => {
    processCount.value = 0;
    try {
        const res = await fileWgetKeys();
        if (res.data && res.data.keys.length > 0) {
            processCount.value = res.data.keys.length;
        }
    } catch (error) {}
};

const openRename = (item: File.File, source: string) => {
    fileRename.path = req.path;
    fileRename.oldName = item.name;
    if (source === 'right') {
        fileRename.newName = item.name;
        isEdit.value = true;
        nextTick(() => {
            getCurrentRename().focus();
        });
        hideRightMenu();
    } else {
        renameRef.value.acceptParams(fileRename);
    }
};

const onRenameBlur = (e: FocusEvent, row: File.File) => {
    const related = e.relatedTarget as HTMLElement | null;
    if (
        related &&
        (related.closest('.fu-table-more-button') || related.closest('.fu-table-more-button .el-dropdown__item'))
    ) {
        setTimeout(() => {
            getCurrentRename()?.focus();
        }, 0);
        return;
    }
    handleRename(row);
};

const handleRename = async (row: File.File): Promise<void> => {
    if (fileRename.newName === fileRename.oldName) {
        isEdit.value = false;
        fileRename.oldName = '';
        return;
    }
    const addItem: File.FileRename = {
        oldName: getPath(fileRename.path, fileRename.oldName),
        newName: getPath(fileRename.path, fileRename.newName),
    };
    loading.value = true;
    try {
        await renameRile(addItem);
        MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
        row.name = fileRename.newName;
        row.path = getPath(req.path, fileRename.newName);
    } catch (error) {
        console.error(error);
    } finally {
        loading.value = false;
        isEdit.value = false;
        fileRename.oldName = '';
    }
};

const getPath = (path: string, name: string) => {
    return path + '/' + name;
};

const openMove = (type: string) => {
    fileMove.type = type;
    fileMove.name = '';
    fileMove.allNames = [];
    fileMove.isDir = false;
    const oldPaths = [];
    for (const s of selects.value) {
        oldPaths.push(s['path']);
    }
    fileMove.count = selects.value.length;
    fileMove.oldPaths = oldPaths;
    if (selects.value.length == 1) {
        fileMove.name = selects.value[0].name;
        fileMove.isDir = selects.value[0].isDir;
    } else {
        const allNames = [];
        for (const s of selects.value) {
            allNames.push(s['name']);
        }
        fileMove.allNames = allNames;
    }
    moveOpen.value = true;
    if (type === 'cut') {
        MsgSuccess(i18n.global.t('file.moveSuccess') + '! ' + i18n.global.t('file.pasteMsg'));
    } else {
        MsgSuccess(i18n.global.t('file.copySuccess') + '! ' + i18n.global.t('file.pasteMsg'));
    }
    updateButtons();
};

const openMoveBtn = (type: string, item: File.File) => {
    selects.value = [];
    selects.value.push(item);
    openMove(type);
};

const closeMove = () => {
    selects.value = [];
    getCurrentTable().clearSelects();
    hideRightMenu();
    fileMove.oldPaths = [];
    fileMove.name = '';
    fileMove.count = 0;
    fileMove.isDir = false;
    moveOpen.value = false;
    updateButtons();
};

const openPaste = () => {
    fileMove.path = req.path;
    moveRef.value.acceptParams(fileMove);
};

function onLoading(isLoading: boolean) {
    loading.value = isLoading;
}

const openDownload = (file: File.File) => {
    downloadFile(file.path, globalStore.currentNode);
};

const fileShareRef = ref<InstanceType<typeof FileShare> | null>(null);
const openShareFile = (row: File.File) => {
    fileShareRef.value?.acceptParams({ path: row.path });
};
const openShareDetail = (path: string) => {
    fileShareRef.value?.acceptParams({ path });
};

const openDetail = (row: File.File) => {
    detailRef.value.acceptParams({ path: row.path });
};

const openRecycleBin = () => {
    recycleBinRef.value.acceptParams();
};

const openFavorite = () => {
    favoriteRef.value.acceptParams();
};

const openShareList = () => {
    shareListRef.value.acceptParams();
};

const openFileHistoryCenter = () => {
    historyDrawerRef.value?.acceptParams({
        path: '',
        content: '',
        language: 'plaintext',
        extension: '',
        dirty: false,
        scope: 'all',
    });
};

const handleFileTaskChange = (type: string, payload: { taskID: string; status: string }) => {
    if (!payload.taskID) {
        return;
    }
    const taskKey = `${type}:${payload.taskID}`;
    const previous = fileTaskStatus[taskKey];
    fileTaskStatus[taskKey] = payload.status;
    if (previous === 'Executing' && payload.status === 'Success') {
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    }
    if (payload.status && payload.status !== 'Executing') {
        delete fileTaskStatus[taskKey];
    }
};

const changeSort = ({ prop, order }) => {
    req.sortBy = prop;
    req.sortOrder = order;
    req.search = '';
    req.page = 1;
    req.pageSize = paginationConfig.pageSize;
    req.containSub = false;
    search();
};

const showFavorite = (row: File.File) => {
    hoveredRowPath.value = row.path;
};

const hideFavorite = () => {
    hoveredRowPath.value = null;
};

const addToFavorite = async (row: File.File) => {
    try {
        await addFavorite(row.path);
        await search();
    } catch (error) {}
};

const remove = async (id: number) => {
    ElMessageBox.confirm(i18n.global.t('file.removeFavorite'), i18n.global.t('commons.msg.remove'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    }).then(async () => {
        try {
            await removeFavorite(id);
            await search();
        } catch (error) {}
    });
};

const getFavorites = async () => {
    try {
        const res = await searchFavorite(req);
        favorites.value = res.data.items;
    } catch (error) {}
};

const removeShareByPath = async (path: string) => {
    ElMessageBox.confirm(i18n.global.t('file.shareCancelConfirm'), i18n.global.t('commons.msg.remove'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    }).then(async () => {
        try {
            await removeFileShare(path);
            await search();
        } catch (error) {}
    });
};

const toFavorite = (row: File.Favorite) => {
    if (row.isDir) {
        jump(row.path);
    } else {
        let file = {} as File.File;
        const extension = getFileExtension(row.name);
        file.path = row.path;
        file.name = row.name;
        file.extension = extension;
        file.mimeType = MimetypeByExtensionObject[extension] || '';
        openView(file);
    }
};

const dialogTerminalRef = ref();
const toTerminal = () => {
    dialogTerminalRef.value!.acceptParams({ cwd: req.path, command: '/bin/sh' });
};

const openWithVSCode = (row: File.File) => {
    dialogVscodeOpenRef.value.acceptParams({ path: row.path + (row.isDir ? '' : ':1:1') });
};

const openRemark = async (row: File.File) => {
    try {
        const res = await ElMessageBox.prompt(i18n.global.t('file.remarkPrompt'), i18n.global.t('file.setRemark'), {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            inputValue: row.remark ?? '',
            inputPlaceholder: i18n.global.t('file.remarkPlaceholder'),
        });
        const remark = res.value ?? '';
        await setFileRemark({ path: row.path, remark: remark });
        row.remark = remark;
        MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
    } catch (error) {
        return;
    }
};

const beforeButtons = [
    {
        label: i18n.global.t('commons.button.open'),
        click: open,
        show: (row: File.File) => {
            return row?.isDir || row?.size <= MAX_OPEN_SIZE || isDecompressFile(row);
        },
    },
    {
        label: i18n.global.t('file.previewLargeFile'),
        click: (row: File.File) => {
            const path = row.isSymlink ? row.linkPath : row.path;
            openTextPreview(path, row.name);
        },
        show: (row: File.File) => {
            return !row?.isDir && row?.size > MAX_OPEN_SIZE && !isDecompressFile(row);
        },
    },
    {
        label: i18n.global.t('commons.button.download'),
        click: (row: File.File) => {
            openDownload(row);
        },
        disabled: (row: File.File) => {
            return row?.isDir;
        },
    },
    {
        label: i18n.global.t('commons.button.copy'),
        click: (row: File.File) => openMoveBtn('copy', row),
    },
    {
        label: i18n.global.t('file.move'),
        click: (row: File.File) => openMoveBtn('cut', row),
    },
    {
        label: i18n.global.t('file.paste'),
        click: openPaste,
        disabled: () => {
            return !moveOpen.value;
        },
    },
    {
        label: i18n.global.t('file.compress'),
        click: (row: File.File) => {
            openCompress([row]);
        },
    },
    {
        label: i18n.global.t('file.deCompress'),
        click: openDeCompress,
        disabled: (row: File.File) => {
            return !isDecompressFile(row);
        },
    },
    {
        label: i18n.global.t('file.editPermissions'),
        click: (row: File.File) => {
            openBatchRole([row]);
        },
    },
    {
        label: i18n.global.t('file.setRemark'),
        hideOnRemarkBlackList: true,
        click: (row: File.File) => {
            openRemark(row);
        },
    },
];
const afterButtons = [
    {
        label: i18n.global.t('commons.button.delete'),
        disabled: (row: File.File) => {
            return row.name == '.1panel_clash';
        },
        click: delFile,
        divided: true,
    },
    {
        label: i18n.global.t('file.copyDir'),
        click: copyDir,
    },
    {
        label: i18n.global.t('file.addFavoriteAction'),
        click: (row: File.File) => {
            addToFavorite(row);
        },
        show: (row: File.File) => row?.favoriteID === 0,
    },
    {
        label: i18n.global.t('file.removeFavoriteAction'),
        click: (row: File.File) => {
            remove(row?.favoriteID);
        },
        show: (row: File.File) => row?.favoriteID > 0,
    },
    {
        label: i18n.global.t('file.shareFile'),
        click: openShareFile,
        show: (row: File.File) => {
            return !row?.isDir && !row?.shareCode;
        },
    },
    {
        label: i18n.global.t('file.shareCancel'),
        click: (row: File.File) => {
            removeShareByPath(row.path);
        },
        show: (row: File.File) => {
            return !row?.isDir && !!row?.shareCode;
        },
    },
    {
        label: i18n.global.t('file.convert'),
        click: (row: File.File) => {
            openConvert(row);
        },
        disabled: (row: File.File) => {
            return row?.isDir || !isConvertible(row?.extension, row?.mimeType);
        },
    },
    {
        label: i18n.global.t('file.openWithVscode'),
        click: openWithVSCode,
    },
    {
        label: i18n.global.t('file.info'),
        click: openDetail,
        divided: true,
    },
];

const rightBtnRename = [
    {
        label: i18n.global.t('file.rename'),
        click: (row: File.File) => {
            openRename(row, 'right');
        },
    },
];
const moreBtnRename = [
    {
        label: i18n.global.t('file.rename'),
        click: (row: File.File) => {
            openRename(row, 'more');
        },
    },
];

const filterRemarkButtons = (buttons: any[]) => {
    if (!isInRemarkBlackList(req.path)) {
        return buttons;
    }
    return buttons.filter((btn) => !btn.hideOnRemarkBlackList);
};

const rightButtons = computed(() => filterRemarkButtons([...beforeButtons, ...rightBtnRename, ...afterButtons]));
const tableMoreButtons = computed(() => filterRemarkButtons([...beforeButtons, ...moreBtnRename, ...afterButtons]));
const openConvert = (item: File.File) => {
    if (!ffmpegExist.value) {
        ElMessageBox.confirm(i18n.global.t('cronjob.library.noSuchApp', ['FFmpeg']), i18n.global.t('file.convert'), {
            confirmButtonText: i18n.global.t('app.toInstall'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
        }).then(() => {
            routerToNameWithQuery('Library', { t: Date.now(), uncached: 'true' });
        });
        return;
    } else {
        if (!isConvertible(item.extension, item.mimeType)) {
            MsgWarning(i18n.global.t('file.fileCanNotConvert'));
            return;
        }
        const fileType = getFileType(item.extension);
        fileConvert.outputPath = req.path;
        fileConvert.files = [
            {
                type: fileType,
                path: req.path,
                extension: item.extension,
                inputFile: item.name,
                outputFormat: item.extension.slice(1),
            },
        ];

        convertRef.value.acceptParams(fileConvert);
    }
};

const isDecompressFile = (row: File.File) => {
    if (row.isDir) {
        return false;
    }

    const extension = getFileExtension(row.name, row.extension);
    const mimeType = row.mimeType || MimetypeByExtensionObject[extension];

    if (getFileType(extension) === 'compress') {
        return true;
    }

    if (!mimeType || mimeType === 'application/octet-stream') {
        return false;
    }

    return Mimetypes.get(mimeType) != undefined;
};

const getHostMount = async () => {
    try {
        const res = await searchHostMount();
        hostMount.value = res.data;
    } catch (error) {
        console.error('Error fetching host mount:', error);
    }
};

const handleDrop = async (event: DragEvent) => {
    event.preventDefault();
    isDragOver.value = false;
    fileUpload.path = req.path;
    if (!uploadRef.value?.open) {
        await uploadRef.value?.handleDrop(event);
    }
    uploadRef.value.acceptParams(fileUpload);
};

const handleDragover = (event: DragEvent) => {
    event.preventDefault();
    isDragOver.value = true;
};

const handleDragleave = (event: { preventDefault: () => void }) => {
    event.preventDefault();
    isDragOver.value = false;
};

function hideRightMenu() {
    getCurrentTable().closeRightClick();
}

function initShowHidden() {
    const showHidden = localStorage.getItem('show-hidden');
    if (showHidden === null) {
        localStorage.setItem('show-hidden', 'true');
        req.showHidden = true;
    } else {
        req.showHidden = showHidden === 'true';
    }
}

const remarkBlackList = ['/proc', '/sys', '/dev', '/run'];

const scheduleRemarkLoad = () => {
    if (remarkLoadTimer.value) {
        window.clearTimeout(remarkLoadTimer.value);
    }
    if (isInRemarkBlackList(req.path)) {
        return;
    }
    remarkLoadTimer.value = window.setTimeout(() => {
        void loadRemarksForCurrentPage();
    }, 1000);
};

const isInRemarkBlackList = (path: string) => {
    return remarkBlackList.some((prefix) => path === prefix || path.startsWith(`${prefix}/`));
};

const loadRemarksForCurrentPage = async () => {
    if (!Array.isArray(data.value) || data.value.length === 0) return;
    const paths = data.value.map((item) => item.path).filter(Boolean);
    if (paths.length === 0) return;
    const currentId = ++remarkRequestId.value;
    try {
        const res = await batchGetFileRemarks(paths);
        if (currentId !== remarkRequestId.value) return;
        const remarks = res.data?.remarks || {};
        data.value.forEach((item) => {
            const remark = remarks[item.path];
            if (remark !== undefined && remark !== '') {
                item.remark = remark;
            }
        });
    } catch (error) {
        if (currentId !== remarkRequestId.value) return;
    }
};

function initTabsAndPaths() {
    initTabs();
    let path = getInitialPath();
    req.path = path;
    getPaths(path);
    updateTab(path);
    paths.value = buildPaths(path);
    resetPaths();
    pathWidth.value = getCurrentPath()?.offsetWidth;
}

function buildPaths(path: string) {
    return path
        .split('/')
        .filter(Boolean)
        .reduce((accumulator, segment) => {
            const lastPath = accumulator[accumulator.length - 1];
            const currentUrl = lastPath ? `${lastPath.url}/${segment}` : `/${segment}`;
            accumulator.push({
                url: currentUrl,
                name: segment,
            });

            return accumulator;
        }, []);
}

async function initHistory() {
    await loadInitialExistingPath(req.path);
    await search();
    history.push(req.path);
    pointer = history.length - 1;
}

function getInitialPath(): string {
    const routePath = router.currentRoute.value.query.path;
    if (routePath && typeof routePath === 'string') {
        const p = routePath.trim();
        if (p !== '') {
            globalStore.lastFilePath = p;
            return p;
        }
    }
    const tab = editableTabs.value.find((t) => t.id === editableTabsKey.value);
    if (tab && typeof tab.path === 'string' && tab.path.trim() !== '') {
        const p = tab.path.trim();
        globalStore.lastFilePath = p;
        return p;
    }
    if (typeof globalStore.lastFilePath === 'string' && globalStore.lastFilePath.trim() !== '') {
        return globalStore.lastFilePath;
    }

    return '/';
}

function initTabs() {
    const savedTabs = localStorage.getItem('editableTabs');
    if (savedTabs) {
        editableTabs.value = JSON.parse(savedTabs);
    }
    const savedTabsKey = localStorage.getItem('editableTabsKey');
    if (savedTabsKey) {
        editableTabsKey.value = savedTabsKey;
    } else {
        setFirstTab();
    }
}

function setFirstTab() {
    if (editableTabs.value.length > 0) {
        const first = editableTabs.value[0];
        editableTabsKey.value = first.id;
    } else {
        initTabs();
    }
}

function saveStorageTabs() {
    localStorage.setItem('editableTabs', JSON.stringify(editableTabs.value));
}

function saveStorageTabsKey() {
    localStorage.setItem('editableTabsKey', editableTabsKey.value);
}

function getLastPath(path: string): string {
    if (!path) return '';
    const parts = path.split('/').filter(Boolean);
    return parts.length ? parts[parts.length - 1] : '';
}

function updateTab(newPath?: string) {
    editableTabs.value = editableTabs.value.map((tab) => {
        if (tab.id === editableTabsKey.value) {
            return {
                ...tab,
                path: newPath,
                name: getLastPath(newPath),
            };
        }
        return tab;
    });
    saveStorageTabs();
}

const loadPath = async () => {
    const pathRes = await loadBaseDir();
    baseDir.value = pathRes.data;
};

const addTab = () => {
    if (editableTabs.value.length >= 6) {
        MsgWarning(i18n.global.t('file.notCanTab'));
        return;
    }
    const usedIds = new Set(editableTabs.value.map((t) => Number(t.id)));
    const newId = Array.from({ length: 6 }, (_, i) => i + 1).find((id) => !usedIds.has(id));

    if (!newId) {
        MsgWarning(i18n.global.t('file.notCanTab'));
        return;
    }
    editableTabs.value.push({
        id: String(newId),
        name: getLastPath(baseDir.value),
        path: baseDir.value,
    });
    editableTabsKey.value = String(newId);
    changeTab(newId);
};

const changeTab = (targetPath: TabPaneName) => {
    if (targetPath === 99) {
        return;
    }
    const current = editableTabs.value.find((tab) => tab.id === targetPath.toString());
    editableTabsKey.value = current.id;
    saveStorageTabs();
    saveStorageTabsKey();
    req.path = current ? current.path : '';
    globalStore.lastFilePath = req.path;
    getPaths(req.path);
    search();
};

const removeTab = (targetId: TabPaneName) => {
    const tabs = editableTabs.value;
    if (tabs.length <= 1) {
        MsgWarning(i18n.global.t('file.keepOneTab'));
        return;
    }
    const target = String(targetId);
    const current = String(editableTabsKey.value);
    const idx = tabs.findIndex((t) => String(t.id) === target);
    if (idx === -1) return;
    let nextActive = current;
    if (current === target) {
        nextActive = tabs[idx + 1]?.id ?? tabs[idx - 1]?.id ?? current;
    }
    editableTabs.value = tabs.filter((t) => String(t.id) !== target);
    editableTabsKey.value = String(nextActive);
    changeTab(nextActive);
};

const checkFFmpeg = () => {
    getComponentInfo('ffmpeg', globalStore.currentNode).then((res) => {
        ffmpegExist.value = res.data.exists ?? false;
    });
};

const updateHeight = () => {
    const el = fileTableRef.value;
    if (!el) return;
    let tabHeight = globalStore.openMenuTabs ? 40 : -4;
    const half = (el.offsetHeight + tabHeight) / 2;
    dropdownMaxHeight.value = Math.max(half, 300);
};

onMounted(async () => {
    await loadPath();
    await nextTick();
    watchTitleHeight();
    window.addEventListener('resize', watchTitleHeight);
    updateHeight();
    window.addEventListener('resize', updateHeight);
    initShowHidden();
    initTabsAndPaths();
    await getHostMount();
    await initHistory();
    checkFFmpeg();
    await nextTick(function () {
        handlePath();
        observeResize();
    });
});

onBeforeUnmount(() => {
    if (resizeObserver) resizeObserver.disconnect();
    window.removeEventListener('resize', watchTitleHeight);
    window.removeEventListener('resize', updateHeight);
    if (remarkLoadTimer.value) {
        window.clearTimeout(remarkLoadTimer.value);
    }
});
</script>

<style scoped lang="scss">
.file-management-page {
    position: relative;

    &::after {
        position: absolute;
        inset: 52px 12px 12px;
        z-index: 20;
        pointer-events: none;
        content: '';
        border: 1px dashed transparent;
        border-radius: 8px;
        transition:
            background-color 0.2s,
            border-color 0.2s;
    }

    &.is-drag-over::after {
        background-color: var(--el-color-primary-light-9);
        border-color: var(--el-color-primary);
    }
}

.file-navigation {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 2px 0 12px;
}

.file-navigation__actions {
    display: flex;
    flex-shrink: 0;
    align-items: center;
    gap: 6px;
    .el-button + .el-button {
        margin-left: 0px;
    }
}

.file-navigation__path {
    flex: 1;
    min-width: 0;
}

.address-input {
    width: 100%;
}

.address-bar {
    display: flex;
    flex-grow: 1;
    align-items: center;
    padding: 4px 8px;
    overflow: hidden;
    cursor: text;
    background-color: var(--el-fill-color-lighter);
    border: 1px solid var(--el-border-color-light);
    border-radius: 6px;
    transition:
        border-color 0.2s,
        box-shadow 0.2s;

    &:hover {
        border-color: var(--el-color-primary-light-5);
        box-shadow: 0 0 0 2px var(--el-color-primary-light-9);
    }

    .arrow {
        color: var(--el-text-color-placeholder);
    }
}

.address-url {
    gap: 2px;
    min-width: 0;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
}

.breadcrumb-root {
    display: inline-flex;
    flex-shrink: 0;
    align-items: center;
    justify-content: center;
    width: 22px;
    height: 24px;
    color: var(--el-text-color-secondary);
    transition: color 0.2s;

    &:hover {
        color: var(--el-color-primary);
    }
}

.breadcrumb-item {
    flex-shrink: 0;

    &:last-child {
        flex-shrink: 1;

        .path-segment {
            max-width: none;
        }
    }
}

.arrow {
    margin: 0 2px;
    font-size: 12px;
    color: var(--el-text-color-placeholder);
}

.path-segment {
    display: inline-block;
    max-width: 180px;
    padding: 0 2px;
    overflow: hidden;
    color: var(--el-text-color-regular);
    text-overflow: ellipsis;
    vertical-align: bottom;
    white-space: nowrap;
    transition: color 0.2s;

    &:hover {
        color: var(--el-color-primary);
    }

    &.path-segment--overflow {
        font-weight: 600;
        color: var(--el-text-color-secondary);
    }
}

.file-left-toolbar,
.file-right-toolbar {
    display: flex;
    align-items: center;
    gap: 8px;
    max-width: 100%;
}

.file-left-toolbar {
    flex-wrap: wrap;
}

.file-utility-group {
    display: inline-flex;
    flex-wrap: wrap;
    gap: 0;
}

.file-right-toolbar {
    justify-content: flex-start;
    flex-wrap: wrap;
    width: 100%;
}

.file-batch-actions {
    display: flex;
    align-items: center;
    order: 1;
    min-width: 0;
}

.file-batch-group {
    display: flex;
    align-items: center;
    flex-wrap: nowrap;
    max-width: 100%;
}

.file-search-actions {
    display: flex;
    align-items: center;
    justify-content: flex-start;
    order: 2;
    gap: 8px;
    min-width: 0;
}

.file-search-input {
    max-width: 310px;
    min-width: 290px;
}

.file-ai-button {
    flex-shrink: 0;
}

.path {
    display: flex;
    align-items: center;
    border: 1px solid #ebeef5;
    background-color: var(--panel-path-bg);
    height: 30px;
    border-radius: 2px !important;
    &:hover {
        cursor: text;
        box-shadow: var(--el-box-shadow);
    }

    .root {
        vertical-align: middle;
        margin-left: 10px;
    }
    .other {
        vertical-align: middle;
    }
    .split {
        margin-left: 5px;
        margin-right: 5px;
    }
}

.btn-container {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;
}

.copy-button {
    .close {
        width: 10px;
    }
}

.favorite-item {
    height: 30vh;
    overflow: auto;
}

.file-row {
    display: grid;
    grid-template-columns: 24px minmax(0, 1fr) auto auto;
    align-items: center;
    width: 100%;
    min-height: 28px;
    column-gap: 6px;
}

.file-row__icon,
.file-row__actions {
    display: inline-flex;
    align-items: center;
    justify-content: center;
}

.file-name {
    min-width: 0;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
}

.file-pagination-summary {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
    flex-wrap: wrap;
}

.search-button {
    width: 20vw;
}
.el-button-group > .el-dropdown > .el-button {
    border-left-color: var(--el-border-color);
}
.table-input {
    --el-input-inner-height: 22px !important;
}
:deep(.file-tabs .el-tabs__nav .el-tabs__item:last-child) {
    border-bottom: 1px solid var(--el-border-color-light) !important;
}

:deep(.file-tabs .el-tabs--card > .el-tabs__header .el-tabs__item.is-active) {
    border-bottom-width: 1px !important;
}
:deep(.file-tabs .el-tabs--card .el-tabs__header .el-tabs__nav) {
    border-bottom: none !important;
}
:deep(.file-tool) {
    padding: 0 !important;
    .el-button.is-link {
        padding: 8px 15px !important;
    }
}
.file-tool:hover {
    color: var(--el-color-primary) !important;
    .el-button {
        color: var(--el-color-primary) !important;
    }
}

:deep(.file-layout .content-container__toolbar) {
    row-gap: 10px;
}

:deep(.file-layout .content-container__title > .flex) {
    width: 100%;
}

:deep(.file-layout .content-container__title > .flex > div:last-child) {
    flex: 0 1 auto;
    margin-left: auto;
    min-width: 0;
}

:deep(.file-layout .content-container__title > .flex > div:last-child:has(.file-batch-toolbar.is-toolbar-wrapped)) {
    flex: 1 1 100%;
    margin-left: 0;
}

:deep(.file-table .el-table__row) {
    cursor: default;
}

:deep(.file-table .el-table__row:hover > td.el-table__cell) {
    background-color: var(--el-fill-color-light);
}

@media (max-width: 1200px) {
    .file-right-toolbar {
        width: 100%;
    }
}

@media (max-width: 768px) {
    .file-navigation {
        align-items: stretch;
        flex-direction: column;
        gap: 8px;
    }

    .file-navigation__actions,
    .file-search-actions,
    .file-batch-actions,
    .copy-button {
        width: 100%;
    }

    .file-right-toolbar {
        align-items: stretch;
        justify-content: stretch;
    }

    .file-search-actions {
        flex-wrap: wrap;
        justify-content: stretch;
    }

    .file-batch-group {
        flex-wrap: wrap;
    }

    .file-search-input,
    .file-ai-button {
        width: 100%;
        min-width: 0;
    }

    .file-batch-actions {
        justify-content: flex-start;
    }
}
</style>
