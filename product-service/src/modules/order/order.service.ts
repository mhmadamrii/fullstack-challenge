import { Inject, Injectable } from '@nestjs/common';
import { PrismaService } from 'src/prisma.service';
import type { RedisClientType } from 'redis';

@Injectable()
export class OrderService {
  constructor(
    private prisma: PrismaService,
    @Inject('REDIS_CLIENT') private redis: RedisClientType,
  ) {}

  async getAllOrders() {
    return this.prisma.orders.findMany();
  }

  async deleteAllOrders() {
    await this.redis.del('orders');
    return this.prisma.orders.deleteMany();
  }
}
