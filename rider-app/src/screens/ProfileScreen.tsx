import React from 'react';
import {
  View,
  Text,
  TouchableOpacity,
  StyleSheet,
  Alert,
  ScrollView,
} from 'react-native';
import { useNavigation } from '@react-navigation/native';
import { useAuthStore } from '../store/authStore';
import { authApi } from '../api/auth';

export const ProfileScreen: React.FC = () => {
  const navigation = useNavigation();
  const { userInfo, logout } = useAuthStore();

  const handleLogout = async () => {
    Alert.alert('确认', '确定要退出登录吗？', [
      { text: '取消', style: 'cancel' },
      {
        text: '确定',
        style: 'destructive',
        onPress: async () => {
          try {
            await authApi.logout(userInfo?.user_id || '');
            await logout();
          } catch (error) {
            console.error('Logout error:', error);
            await logout();
          }
        },
      },
    ]);
  };

  const menuItems = [
    {
      icon: '⭐',
      title: '我的评分',
      subtitle: '查看服务评分详情',
      onPress: () => Alert.alert('提示', '评分详情功能开发中'),
    },
    {
      icon: '📋',
      title: '我的工单',
      subtitle: '查看申诉和反馈记录',
      onPress: () => Alert.alert('提示', '工单功能开发中'),
    },
    {
      icon: '📝',
      title: '资质认证',
      subtitle: '实名认证、健康证等',
      onPress: () => Alert.alert('提示', '资质认证功能开发中'),
    },
    {
      icon: '💰',
      title: '收入明细',
      subtitle: '查看收入和提现记录',
      onPress: () => navigation.navigate('Income' as never),
    },
    {
      icon: '👥',
      title: '邀请好友',
      subtitle: '邀请骑手赚取奖励',
      onPress: () => navigation.navigate('Referral' as never),
    },
    {
      icon: '⚙️',
      title: '设置',
      subtitle: '账号设置、通知设置',
      onPress: () => Alert.alert('提示', '设置功能开发中'),
    },
    {
      icon: '❓',
      title: '帮助中心',
      subtitle: '常见问题、联系客服',
      onPress: () => navigation.navigate('AIAgent' as never),
    },
  ];

  return (
    <ScrollView style={styles.container}>
      <View style={styles.header}>
        <View style={styles.avatar}>
          <Text style={styles.avatarText}>
            {userInfo?.nickname?.charAt(0) || '骑'}
          </Text>
        </View>
        <Text style={styles.nickname}>{userInfo?.nickname || '美团骑手'}</Text>
        <Text style={styles.phone}>{userInfo?.phone || ''}</Text>
      </View>

      <View style={styles.statsRow}>
        <View style={styles.statItem}>
          <Text style={styles.statValue}>0</Text>
          <Text style={styles.statLabel}>今日订单</Text>
        </View>
        <View style={styles.statDivider} />
        <View style={styles.statItem}>
          <Text style={styles.statValue}>0</Text>
          <Text style={styles.statLabel}>本月订单</Text>
        </View>
        <View style={styles.statDivider} />
        <View style={styles.statItem}>
          <Text style={styles.statValue}>5.0</Text>
          <Text style={styles.statLabel}>服务评分</Text>
        </View>
      </View>

      <View style={styles.menuSection}>
        {menuItems.map((item, index) => (
          <TouchableOpacity
            key={index}
            style={styles.menuItem}
            onPress={item.onPress}
          >
            <Text style={styles.menuIcon}>{item.icon}</Text>
            <View style={styles.menuContent}>
              <Text style={styles.menuTitle}>{item.title}</Text>
              <Text style={styles.menuSubtitle}>{item.subtitle}</Text>
            </View>
            <Text style={styles.menuArrow}>›</Text>
          </TouchableOpacity>
        ))}
      </View>

      <TouchableOpacity style={styles.logoutBtn} onPress={handleLogout}>
        <Text style={styles.logoutBtnText}>退出登录</Text>
      </TouchableOpacity>

      <Text style={styles.version}>版本 1.0.0</Text>
    </ScrollView>
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
  avatar: {
    width: 80,
    height: 80,
    borderRadius: 40,
    backgroundColor: '#fff',
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: 12,
  },
  avatarText: {
    fontSize: 32,
    fontWeight: 'bold',
    color: '#FFD700',
  },
  nickname: {
    fontSize: 20,
    fontWeight: 'bold',
    color: '#333',
    marginBottom: 4,
  },
  phone: {
    fontSize: 14,
    color: '#666',
  },
  statsRow: {
    flexDirection: 'row',
    backgroundColor: '#fff',
    paddingVertical: 20,
    marginBottom: 12,
  },
  statItem: {
    flex: 1,
    alignItems: 'center',
  },
  statValue: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#333',
    marginBottom: 4,
  },
  statLabel: {
    fontSize: 12,
    color: '#999',
  },
  statDivider: {
    width: 1,
    backgroundColor: '#eee',
  },
  menuSection: {
    backgroundColor: '#fff',
    marginBottom: 12,
  },
  menuItem: {
    flexDirection: 'row',
    alignItems: 'center',
    padding: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#f5f5f5',
  },
  menuIcon: {
    fontSize: 24,
    marginRight: 12,
  },
  menuContent: {
    flex: 1,
  },
  menuTitle: {
    fontSize: 16,
    color: '#333',
    marginBottom: 2,
  },
  menuSubtitle: {
    fontSize: 12,
    color: '#999',
  },
  menuArrow: {
    fontSize: 20,
    color: '#ccc',
  },
  logoutBtn: {
    backgroundColor: '#fff',
    marginHorizontal: 16,
    paddingVertical: 14,
    borderRadius: 8,
    alignItems: 'center',
    marginBottom: 16,
  },
  logoutBtnText: {
    fontSize: 16,
    color: '#F44336',
  },
  version: {
    textAlign: 'center',
    fontSize: 12,
    color: '#999',
    marginBottom: 20,
  },
});
