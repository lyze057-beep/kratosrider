import React, { useState, useEffect } from 'react';
import {
  View,
  Text,
  FlatList,
  TouchableOpacity,
  StyleSheet,
  RefreshControl,
  Alert,
  Modal,
  TextInput,
  ScrollView,
} from 'react-native';
import { incomeApi, IncomeInfo, WithdrawalInfo, INCOME_TYPE, WITHDRAWAL_STATUS_TEXT } from '../api/income';
import { useAuthStore } from '../store/authStore';

export const IncomeScreen: React.FC = () => {
  const { userInfo } = useAuthStore();
  const [totalIncome, setTotalIncome] = useState(0);
  const [incomes, setIncomes] = useState<IncomeInfo[]>([]);
  const [withdrawals, setWithdrawals] = useState<WithdrawalInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [showWithdrawModal, setShowWithdrawModal] = useState(false);
  const [withdrawAmount, setWithdrawAmount] = useState('');
  const [withdrawPlatform, setWithdrawPlatform] = useState('alipay');
  const [withdrawAccount, setWithdrawAccount] = useState('');
  const [activeTab, setActiveTab] = useState<'income' | 'withdrawal'>('income');

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    setLoading(true);
    try {
      const riderId = Number(userInfo?.user_id || 0);
      
      const [totalRes, incomeRes, withdrawalRes] = await Promise.all([
        incomeApi.getTotalIncome(riderId),
        incomeApi.getIncomeList({ rider_id: riderId, limit: 20 }),
        incomeApi.getWithdrawalList({ rider_id: riderId, limit: 20 }),
      ]);

      setTotalIncome(totalRes.total);
      setIncomes(incomeRes.incomes || []);
      setWithdrawals(withdrawalRes.withdrawals || []);
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

  const handleWithdraw = async () => {
    const amount = parseFloat(withdrawAmount);
    if (!amount || amount <= 0) {
      Alert.alert('提示', '请输入正确的金额');
      return;
    }

    if (amount > totalIncome) {
      Alert.alert('提示', '余额不足');
      return;
    }

    if (!withdrawAccount) {
      Alert.alert('提示', '请输入提现账号');
      return;
    }

    try {
      const result = await incomeApi.applyWithdrawal({
        rider_id: Number(userInfo?.user_id || 0),
        amount,
        platform: withdrawPlatform,
        account: withdrawAccount,
      });

      if (result.success) {
        Alert.alert('成功', '提现申请已提交');
        setShowWithdrawModal(false);
        setWithdrawAmount('');
        setWithdrawAccount('');
        loadData();
      } else {
        Alert.alert('提示', result.message);
      }
    } catch (error: any) {
      Alert.alert('错误', error.response?.data?.message || '提现失败');
    }
  };

  const renderIncomeItem = ({ item }: { item: IncomeInfo }) => (
    <View style={styles.recordCard}>
      <View style={styles.recordLeft}>
        <Text style={styles.recordTitle}>
          {item.income_type === INCOME_TYPE.ORDER ? '订单收入' : '其他收入'}
        </Text>
        <Text style={styles.recordTime}>{item.created_at}</Text>
      </View>
      <Text style={styles.recordAmount}>+¥{item.amount.toFixed(2)}</Text>
    </View>
  );

  const renderWithdrawalItem = ({ item }: { item: WithdrawalInfo }) => (
    <View style={styles.recordCard}>
      <View style={styles.recordLeft}>
        <Text style={styles.recordTitle}>
          {item.platform === 'alipay' ? '支付宝' : item.platform === 'wechat' ? '微信' : '银行卡'}提现
        </Text>
        <Text style={styles.recordTime}>{item.created_at}</Text>
      </View>
      <View style={styles.recordRight}>
        <Text style={styles.withdrawAmount}>-¥{item.amount.toFixed(2)}</Text>
        <Text style={styles.withdrawStatus}>{WITHDRAWAL_STATUS_TEXT[item.status]}</Text>
      </View>
    </View>
  );

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        <Text style={styles.totalLabel}>总收入（元）</Text>
        <Text style={styles.totalAmount}>¥{totalIncome.toFixed(2)}</Text>
        <TouchableOpacity
          style={styles.withdrawBtn}
          onPress={() => setShowWithdrawModal(true)}
        >
          <Text style={styles.withdrawBtnText}>提现</Text>
        </TouchableOpacity>
      </View>

      <View style={styles.tabs}>
        <TouchableOpacity
          style={[styles.tab, activeTab === 'income' && styles.tabActive]}
          onPress={() => setActiveTab('income')}
        >
          <Text style={[styles.tabText, activeTab === 'income' && styles.tabTextActive]}>
            收入明细
          </Text>
        </TouchableOpacity>
        <TouchableOpacity
          style={[styles.tab, activeTab === 'withdrawal' && styles.tabActive]}
          onPress={() => setActiveTab('withdrawal')}
        >
          <Text style={[styles.tabText, activeTab === 'withdrawal' && styles.tabTextActive]}>
            提现记录
          </Text>
        </TouchableOpacity>
      </View>

      <FlatList
        data={activeTab === 'income' ? incomes : withdrawals}
        keyExtractor={(item) => item.id.toString()}
        renderItem={activeTab === 'income' ? renderIncomeItem : renderWithdrawalItem}
        contentContainerStyle={styles.listContent}
        refreshControl={
          <RefreshControl refreshing={refreshing} onRefresh={handleRefresh} />
        }
        ListEmptyComponent={
          <View style={styles.emptyContainer}>
            <Text style={styles.emptyText}>暂无记录</Text>
          </View>
        }
      />

      <Modal
        visible={showWithdrawModal}
        transparent
        animationType="slide"
        onRequestClose={() => setShowWithdrawModal(false)}
      >
        <View style={styles.modalOverlay}>
          <View style={styles.modalContent}>
            <Text style={styles.modalTitle}>申请提现</Text>

            <Text style={styles.inputLabel}>提现金额</Text>
            <TextInput
              style={styles.input}
              placeholder="请输入提现金额"
              keyboardType="decimal-pad"
              value={withdrawAmount}
              onChangeText={setWithdrawAmount}
            />

            <Text style={styles.inputLabel}>提现方式</Text>
            <View style={styles.platformRow}>
              {['alipay', 'wechat', 'bank'].map((platform) => (
                <TouchableOpacity
                  key={platform}
                  style={[
                    styles.platformBtn,
                    withdrawPlatform === platform && styles.platformBtnActive,
                  ]}
                  onPress={() => setWithdrawPlatform(platform)}
                >
                  <Text
                    style={[
                      styles.platformBtnText,
                      withdrawPlatform === platform && styles.platformBtnTextActive,
                    ]}
                  >
                    {platform === 'alipay' ? '支付宝' : platform === 'wechat' ? '微信' : '银行卡'}
                  </Text>
                </TouchableOpacity>
              ))}
            </View>

            <Text style={styles.inputLabel}>提现账号</Text>
            <TextInput
              style={styles.input}
              placeholder="请输入提现账号"
              value={withdrawAccount}
              onChangeText={setWithdrawAccount}
            />

            <View style={styles.modalButtons}>
              <TouchableOpacity
                style={styles.modalCancelBtn}
                onPress={() => setShowWithdrawModal(false)}
              >
                <Text style={styles.modalCancelBtnText}>取消</Text>
              </TouchableOpacity>
              <TouchableOpacity style={styles.modalConfirmBtn} onPress={handleWithdraw}>
                <Text style={styles.modalConfirmBtnText}>确认提现</Text>
              </TouchableOpacity>
            </View>
          </View>
        </View>
      </Modal>
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
  totalLabel: {
    fontSize: 14,
    color: '#666',
    marginBottom: 8,
  },
  totalAmount: {
    fontSize: 36,
    fontWeight: 'bold',
    color: '#333',
    marginBottom: 16,
  },
  withdrawBtn: {
    backgroundColor: '#333',
    paddingHorizontal: 32,
    paddingVertical: 10,
    borderRadius: 20,
  },
  withdrawBtnText: {
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
  recordTitle: {
    fontSize: 16,
    color: '#333',
    marginBottom: 4,
  },
  recordTime: {
    fontSize: 12,
    color: '#999',
  },
  recordAmount: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#4CAF50',
  },
  recordRight: {
    alignItems: 'flex-end',
  },
  withdrawAmount: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#F44336',
    marginBottom: 4,
  },
  withdrawStatus: {
    fontSize: 12,
    color: '#999',
  },
  emptyContainer: {
    alignItems: 'center',
    paddingVertical: 40,
  },
  emptyText: {
    fontSize: 16,
    color: '#999',
  },
  modalOverlay: {
    flex: 1,
    backgroundColor: 'rgba(0,0,0,0.5)',
    justifyContent: 'flex-end',
  },
  modalContent: {
    backgroundColor: '#fff',
    borderTopLeftRadius: 20,
    borderTopRightRadius: 20,
    padding: 24,
  },
  modalTitle: {
    fontSize: 20,
    fontWeight: 'bold',
    textAlign: 'center',
    marginBottom: 24,
  },
  inputLabel: {
    fontSize: 14,
    color: '#666',
    marginBottom: 8,
  },
  input: {
    height: 48,
    borderWidth: 1,
    borderColor: '#ddd',
    borderRadius: 8,
    paddingHorizontal: 16,
    fontSize: 16,
    marginBottom: 16,
  },
  platformRow: {
    flexDirection: 'row',
    marginBottom: 16,
  },
  platformBtn: {
    flex: 1,
    height: 40,
    borderWidth: 1,
    borderColor: '#ddd',
    justifyContent: 'center',
    alignItems: 'center',
    marginHorizontal: 4,
    borderRadius: 8,
  },
  platformBtnActive: {
    borderColor: '#FFD700',
    backgroundColor: '#FFF9E6',
  },
  platformBtnText: {
    fontSize: 14,
    color: '#666',
  },
  platformBtnTextActive: {
    color: '#FFD700',
    fontWeight: 'bold',
  },
  modalButtons: {
    flexDirection: 'row',
    marginTop: 8,
  },
  modalCancelBtn: {
    flex: 1,
    height: 48,
    justifyContent: 'center',
    alignItems: 'center',
    borderWidth: 1,
    borderColor: '#ddd',
    borderRadius: 8,
    marginRight: 8,
  },
  modalCancelBtnText: {
    fontSize: 16,
    color: '#666',
  },
  modalConfirmBtn: {
    flex: 1,
    height: 48,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#FFD700',
    borderRadius: 8,
    marginLeft: 8,
  },
  modalConfirmBtnText: {
    fontSize: 16,
    fontWeight: 'bold',
    color: '#333',
  },
});
