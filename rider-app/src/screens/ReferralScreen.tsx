import React, { useState, useEffect } from 'react';
import {
  View,
  Text,
  FlatList,
  TouchableOpacity,
  StyleSheet,
  RefreshControl,
  Alert,
  Share,
} from 'react-native';
import { referralApi, InviteCodeInfo, InvitedRiderInfo, TaskInfo, INVITED_RIDER_STATUS } from '../api/referral';
import { useAuthStore } from '../store/authStore';

export const ReferralScreen: React.FC = () => {
  const { userInfo } = useAuthStore();
  const [inviteCodeInfo, setInviteCodeInfo] = useState<InviteCodeInfo | null>(null);
  const [invitedRiders, setInvitedRiders] = useState<InvitedRiderInfo[]>([]);
  const [tasks, setTasks] = useState<TaskInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [activeTab, setActiveTab] = useState<'invite' | 'tasks'>('invite');

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    setLoading(true);
    try {
      const riderId = Number(userInfo?.user_id || 0);

      const [infoRes, recordsRes, tasksRes] = await Promise.all([
        referralApi.getMyReferralInfo(riderId),
        referralApi.getInviteRecordList(riderId, 1, 20),
        referralApi.getTaskList(riderId),
      ]);

      if (infoRes.invite_code_info) {
        setInviteCodeInfo(infoRes.invite_code_info);
      }
      setInvitedRiders(recordsRes.records || []);
      setTasks(tasksRes.tasks || []);
    } catch (error) {
      console.error('Load data error:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleRefresh = async () => {
    setRefreshing(true);
    await loadData();
    setRefreshing(false);
  };

  const handleGenerateCode = async () => {
    try {
      const result = await referralApi.generateInviteCode(Number(userInfo?.user_id || 0));
      if (result.invite_code_info) {
        setInviteCodeInfo(result.invite_code_info);
        Alert.alert('成功', '邀请码生成成功');
      }
    } catch (error: any) {
      Alert.alert('错误', error.response?.data?.message || '生成失败');
    }
  };

  const handleShare = async () => {
    if (!inviteCodeInfo) {
      Alert.alert('提示', '请先生成邀请码');
      return;
    }

    try {
      await Share.share({
        message: `【美团骑手邀请】邀请码：${inviteCodeInfo.invite_code}\n加入美团骑手，开启赚钱之旅！\n${inviteCodeInfo.invite_link}`,
      });
    } catch (error) {
      console.error('Share error:', error);
    }
  };

  const getStatusText = (status: number) => {
    switch (status) {
      case INVITED_RIDER_STATUS.REGISTERED:
        return '已注册';
      case INVITED_RIDER_STATUS.FIRST_ORDER:
        return '已完成首单';
      case INVITED_RIDER_STATUS.TASK_COMPLETED:
        return '任务达标';
      case INVITED_RIDER_STATUS.REWARDED:
        return '已奖励';
      default:
        return '未知';
    }
  };

  const renderInvitedRider = ({ item }: { item: InvitedRiderInfo }) => (
    <View style={styles.recordCard}>
      <View style={styles.recordLeft}>
        <Text style={styles.recordName}>{item.nickname}</Text>
        <Text style={styles.recordPhone}>{item.phone}</Text>
      </View>
      <View style={styles.recordRight}>
        <Text style={styles.recordStatus}>{getStatusText(item.status)}</Text>
        {item.reward_amount > 0 && (
          <Text style={styles.recordReward}>+¥{(item.reward_amount / 100).toFixed(2)}</Text>
        )}
      </View>
    </View>
  );

  const renderTask = ({ item }: { item: TaskInfo }) => (
    <View style={styles.taskCard}>
      <View style={styles.taskHeader}>
        <Text style={styles.taskName}>{item.task_name}</Text>
        <Text style={styles.taskReward}>¥{(item.reward_amount / 100).toFixed(2)}</Text>
      </View>
      <Text style={styles.taskDesc}>{item.task_desc}</Text>
      <View style={styles.taskFooter}>
        <Text style={styles.taskTime}>限时{item.time_limit_days}天</Text>
        <TouchableOpacity style={styles.taskBtn}>
          <Text style={styles.taskBtnText}>查看进度</Text>
        </TouchableOpacity>
      </View>
    </View>
  );

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        {inviteCodeInfo ? (
          <>
            <Text style={styles.inviteCode}>{inviteCodeInfo.invite_code}</Text>
            <Text style={styles.inviteStats}>
              累计邀请 {inviteCodeInfo.total_invited} 人 | 有效 {inviteCodeInfo.valid_invited} 人
            </Text>
            <Text style={styles.totalRewards}>
              累计奖励 ¥{(inviteCodeInfo.total_rewards / 100).toFixed(2)}
            </Text>
          </>
        ) : (
          <TouchableOpacity style={styles.generateBtn} onPress={handleGenerateCode}>
            <Text style={styles.generateBtnText}>生成邀请码</Text>
          </TouchableOpacity>
        )}

        <TouchableOpacity style={styles.shareBtn} onPress={handleShare}>
          <Text style={styles.shareBtnText}>分享邀请</Text>
        </TouchableOpacity>
      </View>

      <View style={styles.tabs}>
        <TouchableOpacity
          style={[styles.tab, activeTab === 'invite' && styles.tabActive]}
          onPress={() => setActiveTab('invite')}
        >
          <Text style={[styles.tabText, activeTab === 'invite' && styles.tabTextActive]}>
            邀请记录
          </Text>
        </TouchableOpacity>
        <TouchableOpacity
          style={[styles.tab, activeTab === 'tasks' && styles.tabActive]}
          onPress={() => setActiveTab('tasks')}
        >
          <Text style={[styles.tabText, activeTab === 'tasks' && styles.tabTextActive]}>
            任务列表
          </Text>
        </TouchableOpacity>
      </View>

      <FlatList
        data={activeTab === 'invite' ? invitedRiders : tasks}
        keyExtractor={(item, index) => index.toString()}
        renderItem={activeTab === 'invite' ? renderInvitedRider : renderTask}
        contentContainerStyle={styles.listContent}
        refreshControl={
          <RefreshControl refreshing={refreshing} onRefresh={handleRefresh} />
        }
        ListEmptyComponent={
          <View style={styles.emptyContainer}>
            <Text style={styles.emptyText}>
              {activeTab === 'invite' ? '暂无邀请记录' : '暂无任务'}
            </Text>
          </View>
        }
      />
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f5f5f5',
  },
  header: {
    backgroundColor: '#FFD700',
    padding: 24,
    alignItems: 'center',
  },
  inviteCode: {
    fontSize: 32,
    fontWeight: 'bold',
    color: '#333',
    letterSpacing: 4,
    marginBottom: 8,
  },
  inviteStats: {
    fontSize: 14,
    color: '#666',
    marginBottom: 4,
  },
  totalRewards: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#FF6B00',
    marginBottom: 16,
  },
  generateBtn: {
    backgroundColor: '#333',
    paddingHorizontal: 32,
    paddingVertical: 12,
    borderRadius: 20,
    marginBottom: 16,
  },
  generateBtnText: {
    color: '#FFD700',
    fontSize: 16,
    fontWeight: 'bold',
  },
  shareBtn: {
    backgroundColor: '#333',
    paddingHorizontal: 32,
    paddingVertical: 10,
    borderRadius: 20,
  },
  shareBtnText: {
    color: '#FFD700',
    fontSize: 16,
    fontWeight: 'bold',
  },
  tabs: {
    flexDirection: 'row',
    backgroundColor: '#fff',
    borderBottomWidth: 1,
    borderBottomColor: '#eee',
  },
  tab: {
    flex: 1,
    paddingVertical: 14,
    alignItems: 'center',
  },
  tabActive: {
    borderBottomWidth: 2,
    borderBottomColor: '#FFD700',
  },
  tabText: {
    fontSize: 15,
    color: '#666',
  },
  tabTextActive: {
    color: '#FFD700',
    fontWeight: 'bold',
  },
  listContent: {
    padding: 16,
  },
  recordCard: {
    backgroundColor: '#fff',
    borderRadius: 12,
    padding: 16,
    marginBottom: 12,
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  recordLeft: {
    flex: 1,
  },
  recordName: {
    fontSize: 16,
    color: '#333',
    marginBottom: 4,
  },
  recordPhone: {
    fontSize: 12,
    color: '#999',
  },
  recordRight: {
    alignItems: 'flex-end',
  },
  recordStatus: {
    fontSize: 12,
    color: '#FFD700',
    marginBottom: 4,
  },
  recordReward: {
    fontSize: 16,
    fontWeight: 'bold',
    color: '#4CAF50',
  },
  taskCard: {
    backgroundColor: '#fff',
    borderRadius: 12,
    padding: 16,
    marginBottom: 12,
  },
  taskHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 8,
  },
  taskName: {
    fontSize: 16,
    fontWeight: 'bold',
    color: '#333',
  },
  taskReward: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#FF6B00',
  },
  taskDesc: {
    fontSize: 14,
    color: '#666',
    marginBottom: 12,
  },
  taskFooter: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  taskTime: {
    fontSize: 12,
    color: '#999',
  },
  taskBtn: {
    backgroundColor: '#FFD700',
    paddingHorizontal: 16,
    paddingVertical: 6,
    borderRadius: 12,
  },
  taskBtnText: {
    fontSize: 12,
    fontWeight: 'bold',
    color: '#333',
  },
  emptyContainer: {
    alignItems: 'center',
    paddingVertical: 40,
  },
  emptyText: {
    fontSize: 16,
    color: '#999',
  },
});
