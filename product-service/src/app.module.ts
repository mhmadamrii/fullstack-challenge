import { Module } from '@nestjs/common';
import { AppController } from './app.controller';
import { AppService } from './app.service';
import { ProductModule } from './modules/product/product.module';
import { RedisModule } from './modules/redis/redis.module';
import { OrderModule } from './modules/order/order.module';

@Module({
  imports: [ProductModule, RedisModule, OrderModule],
  controllers: [AppController],
  providers: [AppService],
})
export class AppModule {}
