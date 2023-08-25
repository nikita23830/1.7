package su.metalabs.metabotania.entity;

import baubles.api.BaublesApi;
import cpw.mods.fml.relauncher.Side;
import cpw.mods.fml.relauncher.SideOnly;
import net.minecraft.block.Block;
import net.minecraft.client.Minecraft;
import net.minecraft.client.gui.ScaledResolution;
import net.minecraft.client.renderer.entity.RenderItem;
import net.minecraft.client.renderer.texture.TextureMap;
import net.minecraft.entity.Entity;
import net.minecraft.entity.SharedMonsterAttributes;
import net.minecraft.entity.effect.EntityLightningBolt;
import net.minecraft.entity.player.EntityPlayer;
import net.minecraft.init.Items;
import net.minecraft.inventory.IInventory;
import net.minecraft.item.Item;
import net.minecraft.item.ItemStack;
import net.minecraft.potion.Potion;
import net.minecraft.potion.PotionEffect;
import net.minecraft.tileentity.TileEntityChest;
import net.minecraft.util.*;
import net.minecraft.world.World;
import org.lwjgl.opengl.ARBShaderObjects;
import org.lwjgl.opengl.GL11;
import org.lwjgl.opengl.GL12;
import su.metalabs.metabotania.MetaBotania;
import su.metalabs.metabotania.utils.FlugelExplosition;
import su.metalabs.metabotania.utils.SphereChecker;
import vazkii.botania.api.boss.IBotaniaBossWithShader;
import vazkii.botania.api.internal.ShaderCallback;
import vazkii.botania.common.Botania;
import vazkii.botania.common.core.helper.ItemNBTHelper;
import vazkii.botania.common.entity.EntityDoppleganger;
import vazkii.botania.common.item.ModItems;
import vazkii.botania.common.item.equipment.bauble.ItemFlightTiara;

import java.util.Iterator;
import java.util.List;

import static vazkii.botania.common.entity.EntityDoppleganger.isTruePlayer;

public class EntityFlugel extends EntityExMachine implements IBotaniaBossWithShader {

    public EntityFlugel(World world, int type) {
        super(world, type);
    }

    public EntityFlugel(World world) {
        super(world);
    }

    public String getCommandSenderName() {
        return StatCollector.translateToLocal("entity.Botania.botania:flugel" + type + ".name");
    }

    public static Boolean checkPylon(World world, int xs, int ys, int zs, EntityPlayer player, int type) {
        int radius = MetaBotania.config.getRadius_check_flugel()[type];
        if (!SphereChecker.isSphereAir(world, xs, ys, zs, radius,
                new int[][]{{xs, ys, zs},
                        {xs, ys - 1, zs},
                        {xs, ys - 1, zs + 1},
                        {xs + 1, ys - 1, zs + 1},
                        {xs + 1, ys - 1, zs},
                        {xs + 1, ys - 1, zs - 1},
                        {xs, ys - 1, zs - 1},
                        {xs - 1, ys - 1, zs - 1},
                        {xs - 1, ys - 1, zs},
                        {xs - 1, ys - 1, zs + 1},
        })) {
            if(!world.isRemote) {
                player.addChatMessage((new ChatComponentTranslation("botaniamisc.obstructed", new Object[0])).setChatStyle((new ChatStyle()).setColor(EnumChatFormatting.RED)));
            }
            return false;
        }
        return true;
    }

    public static boolean hasTiara(EntityPlayer player) {
        IInventory inventory = BaublesApi.getBaubles(player);
        for (int i = 0; i < inventory.getSizeInventory(); ++i) {
            ItemStack stack = inventory.getStackInSlot(i);
            if (stack != null && stack.getItem() == ModItems.flightTiara) {
                return true;
            }
        }
        return false;
    }

    public static boolean spawn(EntityPlayer player, ItemStack item, World world, int xs, int ys, int zs, int type) {
        if (world.isRemote)
            return true;
        if (!checkPylon(world, xs, ys, zs, player, type))
            return false;

        int playerCount = 0;
        EntityFlugel entity = new EntityFlugel(world, type);
        entity.setPosition((double) xs + 0.5D, (double) (ys + 3), (double) zs + 0.5D);
        entity.setInvulTime(0);
        entity.setHealth(entity.getMaxHealth());
        entity.setSource(xs, ys, zs);
        entity.setSource(xs, ys, zs);
        entity.setMobSpawnTicks(900);
        entity.setType(type);
        entity.setHardMode(true);

        List playersAround = entity.getPlayersAround();
        Iterator var24 = playersAround.iterator();
        while (var24.hasNext()) {
            EntityPlayer player_around = (EntityPlayer) var24.next();
            if (isTruePlayer(player_around)) {
                ++playerCount;
                entity.changeListAtacker(player_around);
            }
        }
        entity.setPlayerCount(playerCount);
        entity.getAttributeMap()
                .getAttributeInstance(SharedMonsterAttributes.maxHealth)
                .setBaseValue((double) (MetaBotania.config.getBase_hp_flugel()[type] * (MetaBotania.config.getHp_per_player_flugel()[type] * playerCount)));
        world.playSoundAtEntity(entity, "mob.enderdragon.growl", 10.0F, 0.1F);
        world.spawnEntityInWorld(entity);

        --item.stackSize;
        if (item.stackSize == 0)
            item = null;
        return true;
    }

    protected void applyEntityAttributes() {
        super.applyEntityAttributes();
        this.getEntityAttribute(SharedMonsterAttributes.movementSpeed).setBaseValue(0.4);
        this.getEntityAttribute(SharedMonsterAttributes.maxHealth).setBaseValue((Double) MetaBotania.config.getBase_hp_flugel()[type]);
        this.getEntityAttribute(SharedMonsterAttributes.knockbackResistance).setBaseValue(MetaBotania.config.getKnockback_resistance_flugel()[type]);
    }

    public int getTotalArmorValue() {
        return (int) (super.getTotalArmorValue() + MetaBotania.config.getArmor_flugel()[type]);
    }

    public int getMaxDamageFromPlayer(boolean crit) {
        String dmg = MetaBotania.config.getDamage_fligel()[type];
        return Integer.parseInt(dmg.split(":")[crit ? 1 : 0]);
    }

    @SideOnly(Side.CLIENT)
    public void bossBarRenderCallback(ScaledResolution res, int x, int y) {
        GL11.glPushMatrix();
        int px = x + 160;
        int py = y + 12;

        Minecraft mc = Minecraft.getMinecraft();
        ItemStack stack = new ItemStack(Items.skull, 1, 3);
        mc.renderEngine.bindTexture(TextureMap.locationItemsTexture);
        net.minecraft.client.renderer.RenderHelper.enableGUIStandardItemLighting();
        GL11.glEnable(GL12.GL_RESCALE_NORMAL);
        RenderItem.getInstance().renderItemIntoGUI(mc.fontRenderer, mc.renderEngine, stack, px, py);
        net.minecraft.client.renderer.RenderHelper.disableStandardItemLighting();

        boolean unicode = mc.fontRenderer.getUnicodeFlag();
        mc.fontRenderer.setUnicodeFlag(true);
        mc.fontRenderer.drawStringWithShadow(Integer.toString(getPlayerCount()), px + 15, py + 4, 0xFFFFFF);

        mc.fontRenderer.drawStringWithShadow("6th race in the Ixceed", x, py + 4, 0xFFFFFF);

        mc.fontRenderer.setUnicodeFlag(unicode);
        GL11.glPopMatrix();
    }

    public void addLoot(TileEntityChest chest) {
        if (MetaBotania.config.getCan_drop_base_items_flugel()[type]) {
            EntityDoppleganger fake = new EntityDoppleganger(worldObj);
            fake.setHardMode(true);
            ((IGaia)fake).getPlayerList().addAll(playersWhoAttacked);
            fake.setPlayerCount(playersWhoAttacked.size());
            ((IGaia)fake).dropFewItems();
        }
        int rand = this.rand.nextInt(100);
        for (String drop : MetaBotania.config.getDrop_flugel()[type]) {
            String[] dropData = drop.split(" ");
            int count = 1;
            if (dropData[2].contains("-")) {
                count = Integer.parseInt(dropData[2].split("-")[0]) + this.rand.nextInt(Integer.parseInt(dropData[2].split("-")[1]) - Integer.parseInt(dropData[2].split("-")[0]));
            }
            int meta = Integer.parseInt(dropData[1]);
            int chance = Integer.parseInt(dropData[3]);
            if (chance >= rand) {
                ItemStack item = new ItemStack((Item) Item.itemRegistry.getObject(dropData[0]), count, meta);
                putItemInChest(item, chest);
            }
        }
    }

    @Override
    protected void teleport() {
        this.CD_TP = 100;
        if (!worldObj.isRemote) {
            ChunkCoordinates newCoord = SphereChecker.getRandomEmptyPositionInSphere(worldObj, (int) posX, (int) posY, (int) posZ, MetaBotania.config.getRadius_check_flugel()[type]);
            if (newCoord == null)
                return;
            this.setPosition(newCoord.posX, newCoord.posY, newCoord.posZ);
        }
        if (this.worldObj.isRemote) {
            for(int i = 0; i < 50; i++)
                Botania.proxy.sparkleFX(
                        worldObj,
                        this.posX + Math.random() * this.width,
                        this.posY - 1.6 + Math.random() * this.height,
                        this.posZ + Math.random() * this.width,
                        0.25F, 1F, 0.25F, 1F, 10
                );
        }
    }

    @Override
    protected void attack() {
        if (worldObj.isRemote)
            return;
        int type = this.rand.nextInt(6);
        switch (type) {
            case 0: {
                List pls = getPlayersAround();
                if (pls == null || pls.size() == 0)
                    return;
                EntityPlayer player = (EntityPlayer) pls.get(this.rand.nextInt(pls.size()));
                Vec3 vec3 = player.getLookVec();
                double xc = vec3.xCoord*((double)-1)+(double)player.posX;
                double yc = vec3.yCoord*((double)-1)+(double)player.posY;
                double zc = vec3.zCoord*((double)-1)+(double)player.posZ;
                double xxc = this.posX;
                double yyc = this.posY;
                double zzc = this.posZ;
                this.setPositionAndUpdate(xc, yc, zc);
                player.setPositionAndUpdate(xxc,yyc,zzc);
                player.attackEntityFrom(MetaBotania.ABSOLUTE_DAMAGE, 5F);
                break;
            }
            case 1: {
                List pls = getPlayersAround();
                if (pls == null || pls.size() == 0)
                    return;
                pls.forEach(pl -> {
                    EntityPlayer p = (EntityPlayer) pl;
                    p.addPotionEffect(new PotionEffect(Potion.moveSlowdown.id, 100, 100));
                    p.addPotionEffect(new PotionEffect(Potion.digSlowdown.id, 100, 100));
                    p.addPotionEffect(new PotionEffect(Potion.weakness.id, 100, 100));
                    worldObj.addWeatherEffect(new EntityLightningBolt(worldObj, p.posX, p.posY, p.posZ));
                });
                break;
            }
            case 2: {
                List pls = getPlayersAround();
                if (pls == null || pls.size() == 0)
                    return;
                EntityMagicLandmineII land = new EntityMagicLandmineII(super.worldObj);
                int xc = this.getSource().posX - 10 + super.rand.nextInt(20);
                int zc = this.getSource().posZ - 10 + super.rand.nextInt(20);
                int yc = super.worldObj.getTopSolidOrLiquidBlock(xc, zc);
                land.setPosition((double)xc + 0.5D, (double)yc, (double)zc + 0.5D);
                land.summoner2 = this;
                super.worldObj.spawnEntityInWorld(land);
                this.spawnMissile();
                break;
            }
            case 3: {
                this.setPositionAndUpdate(this.getSource().posX + 0.5D, this.getSource().posY + 2.0D, this.getSource().posZ + 0.5D);
                this.heal(((Double)MetaBotania.config.getHeal_flugel()[this.type]).floatValue());
                break;
            }
            case 4: {
                FlugelExplosition exp = new FlugelExplosition(
                        worldObj,
                        this,
                        (this.getSource().posX + 0.5f),
                        (this.getSource().posY + 0.5f),
                        (this.getSource().posZ + 0.5f),
                        MetaBotania.config.getRadius_check_flugel()[this.type] / 2f
                );
                exp.doExplosionA();
                break;
            }
            case 5: {
                List pls = getPlayersAround();
                if (pls == null || pls.size() == 0)
                    return;
                pls.forEach(pl -> {
                    EntityPlayer p = (EntityPlayer) pl;
                    EntitySpear weapon = new EntitySpear(this.worldObj, p);
                    weapon.setPosition(((Entity)p).posX, ((Entity)p).posY + 5 + 2 * rand.nextInt(10), ((Entity)p).posZ);
                    weapon.setDelay((int)(rand.nextInt(55) * 1.2F));
                    this.worldObj.spawnEntityInWorld(weapon);
                    this.worldObj.playSoundAtEntity(weapon, "botania:babylonSpawn", 1F, 1F + this.worldObj.rand.nextFloat() * 3F);
                });
                break;
            }
        }
    }

    @Override
    protected void finishTime(List<EntityPlayer> pls) {
        this.setInvulTime(101);
    }

    @Override
    public void onLivingUpdate() {
        super.onLivingUpdate();
        super.motionX = 0.0D;
        super.motionY = 0.0D;
        super.motionZ = 0.0D; // no motion
        getPlayersAround().forEach(p -> {
            EntityPlayer pl = (EntityPlayer) p;
            ItemStack slot = getSlotTiara(pl);
            if (slot != null) {
                ItemNBTHelper.setInt(slot, "timeLeft", 1200);
            }
        });
    }

    protected ItemStack getSlotTiara(EntityPlayer player) {
        IInventory inv = BaublesApi.getBaubles(player);
        for (int i = 0; i < inv.getSizeInventory(); i++) {
            ItemStack stack = inv.getStackInSlot(i);
            if (stack != null && stack.getItem() instanceof ItemFlightTiara) {
                return stack;
            }
        }
        return null;
    }
}
